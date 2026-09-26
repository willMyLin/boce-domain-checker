package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "BOCE 检测工具",
		Width:  1080,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 239, G: 239, B: 239, A: 1},
		Menu:             app.BuildMenu(),
		OnStartup:        app.startup,
		// OnBeforeClose 只做轻量拦截（检测中阻止关闭），返回 true = 阻止关闭。
		// 【绝不在这里弹框】macOS/Linux 在 goroutine 里异步调用它，弹框时序不可控；
		// 确认框放在菜单「退出」的 Click 里（menu.go 的 onQuit）。
		OnBeforeClose: func(ctx context.Context) bool {
			return app.preventClose()
		},
		// Windows/Linux 上 textarea 复制粘贴的唯一兜底（WebView 默认右键菜单）。
		EnableDefaultContextMenu: true,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "boce-tool-app-8f1c2d3e",
			OnSecondInstanceLaunch: func(second options.SecondInstanceData) {
				if app.ctx == nil {
					return
				}
				runtime.WindowShow(app.ctx)
				runtime.WindowSetAlwaysOnTop(app.ctx, true)
				runtime.WindowSetAlwaysOnTop(app.ctx, false)
			},
		},
		// 不设 About 的话 macOS 应用菜单里连「About」项都没有。
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   "BOCE 检测工具",
				Message: "版本 " + Version,
			},
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
