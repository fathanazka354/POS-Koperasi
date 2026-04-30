package main

import (
	_ "embed"

	"go.uber.org/fx"
)

//go:embed chat_demo.html
var chatDemoHTML []byte

//go:embed shop_demo.html
var shopDemoHTML []byte

func main() {
	fx.New(fxModule()).Run()
}
