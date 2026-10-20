package main

import (
	"budeku/iam/internal/bootstrap"
	"budeku/iam/internal/bootstrap/data"
	"budeku/iam/internal/conf"
	"budeku/iam/internal/infra/database"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

func main() {
	flag.BoolVar(&conf.Debug, "debug", false, "调试模式")

	var (
		isInit   = flag.Bool("init", false, "初始化")
		username = flag.String("username", "", "管理员用户名")
		email    = flag.String("email", "", "管理员邮箱")
	)

	flag.Parse()

	bootstrap.Setup()
	defer bootstrap.Release()

	// 初始化
	if *isInit {
		if *username == "" || *email == "" {
			log.Fatal("执行 --init 时 --username、--email 不可为空")
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

		return // 不启动服务
	}

	bootstrap.Run()
}
