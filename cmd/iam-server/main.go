package main

import (
	"budeku/iam/internal/bootstrap"
	"budeku/iam/internal/conf"
	"flag"
)

func main() {
	flag.BoolVar(&conf.Debug, "debug", false, "调试模式")

	flag.Parse()

	bootstrap.Run()
}
