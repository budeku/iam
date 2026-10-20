package main

import (
	"budeku/iam/internal/bootstrap"
	"budeku/iam/internal/bootstrap/data"
	"budeku/iam/internal/conf"
	"budeku/iam/internal/infra/database"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

const daemonFile = "daemon.pid"

func getPIDFilePath() string {
	_ = os.MkdirAll(conf.DataDir, 0755)
	return filepath.Join(conf.DataDir, daemonFile)
}

func writePID(pid int) error {
	return os.WriteFile(getPIDFilePath(), []byte(strconv.Itoa(pid)), 0644)
}

func removePID() {
	_ = os.Remove(getPIDFilePath())
}

func getRunningPID() (int, error) {
	dataContent, err := os.ReadFile(getPIDFilePath())
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(dataContent)))
	if err != nil {
		return 0, err
	}

	if isProcessRunning(pid) {
		return pid, nil
	}
	return 0, errors.New("进程未运行")
}

func isProcessRunning(pid int) bool {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH")
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(output), fmt.Sprintf(`"%d"`, pid))
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	return proc.Signal(syscall.Signal(0)) == nil
}

func killProcess(pid int) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
		return cmd.Run()
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return proc.Signal(syscall.SIGTERM)
}

func printUsage() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Println("用法:")
	fmt.Println("  iam-server [command] [arguments]")
	fmt.Println("  iam-server [flags]              (直接前台运行服务)")
	fmt.Println()
	fmt.Println("可用命令:")

	fmt.Fprintln(w, "  init\t初始化系统")
	fmt.Fprintln(w, "  start\t在后台启动服务")
	fmt.Fprintln(w, "  server\t在前台启动服务（供后台调用或手动调试）")
	fmt.Fprintln(w, "  stop\t停止后台服务")
	fmt.Fprintln(w, "  restart\t重启后台服务")
	fmt.Fprintln(w, "  help\t显示帮助信息")
	w.Flush()

	fmt.Println()
	fmt.Println("使用 'iam-server <command> -h' 查看特定命令的帮助。")
}

// 打印特定命令的帮助信息
func printCommandUsage(command string) {
	switch command {
	case "init":
		fmt.Println("用法: iam-server init [options]")
		fmt.Println("\n选项:")
		fmt.Println("  --username string  管理员用户名 (必填)")
		fmt.Println("  --email string     管理员邮箱 (必填)")
		fmt.Println("  --debug            调试模式")
	case "start":
		fmt.Println("用法: iam-server start [options]")
		fmt.Println("\n选项:")
		fmt.Println("  --debug            调试模式")
	case "server":
		fmt.Println("用法: iam-server server [options]")
		fmt.Println("\n选项:")
		fmt.Println("  --debug            调试模式")
	case "stop":
		fmt.Println("用法: iam-server stop")
		fmt.Println("\n停止后台运行的 IAM 服务。")
	case "restart":
		fmt.Println("用法: iam-server restart [options]")
		fmt.Println("\n选项:")
		fmt.Println("  --debug            调试模式")
	default:
		printUsage()
	}
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func handleInit(args []string) {
	if hasHelpFlag(args) {
		printCommandUsage("init")
		return
	}

	initFlags := flag.NewFlagSet("init", flag.ExitOnError)
	debug := initFlags.Bool("debug", false, "调试模式")
	username := initFlags.String("username", "", "管理员用户名")
	email := initFlags.String("email", "", "管理员邮箱")

	initFlags.Parse(args)
	conf.Debug = *debug

	if *username == "" || *email == "" {
		log.Fatal("执行 init 时 --username、--email 不可为空")
	}

	if !usernameRegex.MatchString(*username) {
		log.Fatal("用户名格式不正确（仅支持字母、数字和下划线）")
	}

	cleanUsername := strings.ToLower(strings.TrimSpace(*username))
	cleanEmail := strings.ToLower(strings.TrimSpace(*email))

	if len(cleanUsername) > 32 {
		log.Fatal("用户名长度不能超过 32 个字符")
	}

	if len(cleanEmail) < 6 || len(cleanEmail) > 128 {
		log.Fatal("邮箱长度必须在 6 到 128 个字符之间")
	}

	if !emailRegex.MatchString(cleanEmail) {
		log.Fatal("邮箱格式不正确")
	}

	bootstrap.Setup()
	defer bootstrap.Release()

	result, err := data.Init(context.Background(), database.DB(), &data.Options{
		Username: cleanUsername,
		Email:    cleanEmail,
	})

	if err != nil {
		log.Fatal(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	hasNew := false

	if result.AppCreated {
		fmt.Fprintln(w, "  [新建]\tIAM 应用 ID: client_id =", result.AppClientID)
		hasNew = true
	}

	if result.KeypairCreated {
		fmt.Fprintln(w, "  [新建]\tIAM 签名密钥: kid =", result.KeypairKid)
		hasNew = true
	}

	if result.UserCreated {
		fmt.Fprintln(w, "  [新建]\t已创建系统用户 — 请立即保存密码，后续将不再显示:")
		fmt.Fprintln(w, "\t用户名:\t", result.Username)
		fmt.Fprintln(w, "\t邮箱:\t", result.Email)
		fmt.Fprintln(w, "\t密码:\t", result.Password)
		hasNew = true
	}

	if hasNew {
		w.Flush()
	} else {
		fmt.Println("  [提示] 所有组件均已存在，无需重复初始化。")
	}
}

func handleServer(args []string) {
	if hasHelpFlag(args) {
		printCommandUsage("server")
		return
	}

	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)
	debug := serverFlags.Bool("debug", false, "调试模式")
	serverFlags.Parse(args)
	conf.Debug = *debug

	bootstrap.Setup()
	defer bootstrap.Release()

	bootstrap.Run()
}

func handleStart(args []string) {
	if hasHelpFlag(args) {
		printCommandUsage("start")
		return
	}

	if pid, err := getRunningPID(); err == nil {
		fmt.Printf("服务已经在运行中 (PID: %d)\n", pid)
		return
	}

	executable, err := os.Executable()
	if err != nil {
		log.Fatalf("获取可执行文件路径失败: %v", err)
	}

	cmdArgs := append([]string{"server"}, args...)
	cmd := exec.Command(executable, cmdArgs...)

	if err := cmd.Start(); err != nil {
		log.Fatalf("启动后台服务失败: %v", err)
	}

	if err := writePID(cmd.Process.Pid); err != nil {
		log.Fatalf("写入 PID 文件失败: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	if !isProcessRunning(cmd.Process.Pid) {
		removePID()
		log.Fatalf("服务启动失败，进程已异常退出")
	}

	fmt.Printf("服务已启动 (PID: %d)\n", cmd.Process.Pid)
}

func handleStop(args []string) {
	if hasHelpFlag(args) {
		printCommandUsage("stop")
		return
	}

	pid, err := getRunningPID()
	if err != nil {
		fmt.Println("服务未运行")
		return
	}

	_ = killProcess(pid)

	for range 20 {
		if !isProcessRunning(pid) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	removePID()
	fmt.Println("服务已停止")
}

func handleRestart(args []string) {
	if hasHelpFlag(args) {
		printCommandUsage("restart")
		return
	}

	handleStop(nil)
	time.Sleep(1 * time.Second)
	handleStart(args)
}

func main() {
	if len(os.Args) < 2 {
		handleServer(nil)
		return
	}

	subcommand := os.Args[1]

	if strings.HasPrefix(subcommand, "-") {
		if subcommand == "-h" || subcommand == "--help" {
			printUsage()
			return
		}
		handleServer(os.Args[1:])
		return
	}

	switch subcommand {
	case "init":
		handleInit(os.Args[2:])
	case "start":
		handleStart(os.Args[2:])
	case "server":
		handleServer(os.Args[2:])
	case "stop":
		handleStop(os.Args[2:])
	case "restart":
		handleRestart(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}
