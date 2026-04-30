package main

import (
	_ "embed"

	"go.uber.org/fx"
)

//go:embed chat_demo.html
var chatDemoHTML []byte

func main() {
	fx.New(fxModule()).Run()
}
