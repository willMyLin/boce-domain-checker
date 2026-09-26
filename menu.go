package main

import (
	"fmt"
	"runtime"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	wailsrt "github.com/wailsapp/wails/v2/pkg/runtime"
)

// 菜单动作名 —— 通过 runtime.EventsEmit(ctx, "menu:action", name) 交给前端，
// 与 frontend/src/App.vue 的 handleMenuAction 分发表逐条对应。
const (
	menuActionImportTXT    = "import-txt"
	menuActionExportExcel  = "export-excel"
	menuActionStart        = "start"
	menuActionStop         = "stop"
	menuActionCopyRows     = "copy-rows"
	menuActionClearTargets = "clear-targets"
	menuActionClearRows    = "clear-rows"
	menuActionCloseBlocked = "close-blocked"
)

const (
	boceHomeURL = "https://www.boce.com"
	// BOCE 的 API 说明页（官网导航里的「API接口」），申请 / 查看 API Key 从这里进。
	boceAPIURL = "https://www.boce.com/api.html"

	aboutDialogTitle  = "关于 BOCE 检测工具"
	quitDialogTitle   = "确认退出"
	quitConfirmOK     = "退出"
	quitConfirmCancel = "取消"
)

// BuildMenu 构建应用菜单。
//
// 【硬约束 · Wails v2.16.0】
//  1. mac.go 的默认 Mac 菜单构造函数，以及 menuroles.go 里 About/Quit/Undo/Cut 等
//     Role helper，在 v2.16.0 全被包在注释里不可用，仅
//     AppMenu/EditMenu/WindowMenu 三个存活；自定义项只能用 Label/Text/Separator/
//     Radio/Checkbox/SubMenu 构造。
//  2. Role 项在 Windows/Linux 被静默丢弃，Windows 顶层还会多出一个点不开的空白标题菜单
//     （windows/menu.go 对每个顶层项无条件 AddSubMenu，Role 项 Label 是空串）。
//     ⇒ 顶层只能在 darwin 放 Role 项，windows/linux 一个都不能放。
//  3. macOS 漏放那 3 个 Role 项会丢掉整个应用菜单（含 ⌘Q / ⌘V / ⌘M），必须显式放上。
//     Role 生成的子菜单是封闭的（darwin/menu.go 遇 Role 直接 AppendRole 后返回），
//     自定义项必须放进自己建的 menu.SubMenu 里。
func (a *App) BuildMenu() *menu.Menu {
	if runtime.GOOS == "darwin" {
		return menu.NewMenuFromItems(
			menu.AppMenu(),
			menu.SubMenu("文件", a.buildDarwinFileMenu()),
			menu.EditMenu(),
			menu.SubMenu("检测", a.buildDetectMenu(true)),
			menu.WindowMenu(),
			menu.SubMenu("帮助", a.buildHelpMenu()),
		)
	}

	// windows / linux：零 Role 项。
	return menu.NewMenuFromItems(
		menu.SubMenu("文件", a.buildDesktopFileMenu()),
		menu.SubMenu("编辑", a.buildDesktopEditMenu()),
		menu.SubMenu("检测", a.buildDetectMenu(false)),
		menu.SubMenu("窗口", a.buildDesktopWindowMenu()),
		menu.SubMenu("帮助", a.buildHelpMenu()),
	)
}

// buildDarwinFileMenu —— darwin 的「文件」。
// 「退出」不挂加速键：⌘Q 已被 menu.AppMenu() 的 Role「Quit」占用，再挂会冲突；
// 该 Role 项无法挂 Click（Role 子菜单封闭），所以 ⌘Q 只走 OnBeforeClose 的轻量拦截，
// 带确认框的退出走本菜单项。
func (a *App) buildDarwinFileMenu() *menu.Menu {
	return menu.NewMenuFromItems(
		menu.Text("导入 TXT…", keys.CmdOrCtrl("o"), a.act(menuActionImportTXT)),
		menu.Text("导出 Excel…", keys.CmdOrCtrl("s"), a.act(menuActionExportExcel)),
		menu.Separator(),
		menu.Text("打开 BOCE 官网", keys.Combo("b", keys.CmdOrCtrlKey, keys.ShiftKey), a.openURL(boceHomeURL)),
		menu.Separator(),
		menu.Text("退出", nil, a.onQuit),
	)
}

// buildDesktopFileMenu —— windows / linux 的「文件」。
// Ctrl+Q 在 Windows 加速键白名单内（A-Z），不会静默失效，也不属于约束 4 的编辑类按键。
func (a *App) buildDesktopFileMenu() *menu.Menu {
	return menu.NewMenuFromItems(
		menu.Text("导入 TXT…", keys.CmdOrCtrl("o"), a.act(menuActionImportTXT)),
		menu.Text("导出 Excel…", keys.CmdOrCtrl("s"), a.act(menuActionExportExcel)),
		menu.Separator(),
		menu.Text("退出", keys.CmdOrCtrl("q"), a.onQuit),
	)
}

// buildDesktopEditMenu —— windows / linux 的「编辑」。
// 【约束 4】不放撤销/剪切/复制/粘贴/全选，且这些项一律 accelerator 传 nil：
// Windows 的 AcceleratorKeyCallback 会把按键原样 PostMessage 给 WebView2 的同时
// 再触发菜单快捷键（双触发）；Go 侧 runtime 也读不到 WebView 选区，无法自建复制粘贴，
// 交给 WebView + main.go 的 EnableDefaultContextMenu 兜底。
// 因此本菜单里凡键名为 c/v/x/a/z 的项一律不挂加速键（含「复制结果为文本」）。
func (a *App) buildDesktopEditMenu() *menu.Menu {
	return menu.NewMenuFromItems(
		menu.Text("复制结果为文本", nil, a.act(menuActionCopyRows)),
		menu.Text("清空检测目标", nil, a.act(menuActionClearTargets)),
		menu.Text("清空结果", nil, a.act(menuActionClearRows)),
	)
}

// buildDetectMenu —— 「检测」。
// darwin 为 true 时给「复制结果为文本」挂 ⌘⇧C（约束 4 只针对 Windows/Linux 的双触发）。
//
// 开始检测的加速键按平台分开写，不能照抄 keys.Parse("CmdOrCtrl+Enter")：
// keys.Parse 把 Enter 归一成命名键 "enter"，Windows 白名单认 "ENTER" 没问题，但
// macOS 是把 Accelerator.Key 原样当 AppKit keyEquivalent 用的，必须是单字符，
// "enter" 匹配不到任何按键。darwin 这里用 "\r"（Return 的 keyEquivalent）。
func (a *App) buildDetectMenu(darwin bool) *menu.Menu {
	startAccel := keys.Key("f5")
	copyAccel := (*keys.Accelerator)(nil)
	if darwin {
		startAccel = keys.CmdOrCtrl("\r")
		copyAccel = keys.Combo("c", keys.CmdOrCtrlKey, keys.ShiftKey)
	}

	return menu.NewMenuFromItems(
		menu.Text("开始检测", startAccel, a.act(menuActionStart)),
		menu.Text("停止", keys.CmdOrCtrl("."), a.act(menuActionStop)),
		menu.Separator(),
		menu.Text("复制结果为文本", copyAccel, a.act(menuActionCopyRows)),
	)
}

// buildDesktopWindowMenu —— windows / linux 的「窗口」。
// darwin 的最小化 / 缩放 / 全屏由 menu.WindowMenu() 这个 Role 提供，不能重复放。
func (a *App) buildDesktopWindowMenu() *menu.Menu {
	return menu.NewMenuFromItems(
		menu.Text("最小化", nil, a.onMinimize),
		menu.Text("最大化 / 还原", nil, a.onToggleMaximise),
		menu.Text("全屏", keys.Key("f11"), a.onToggleFullscreen),
		menu.Checkbox("置顶", false, nil, a.onToggleAlwaysOnTop),
	)
}

func (a *App) buildHelpMenu() *menu.Menu {
	return menu.NewMenuFromItems(
		menu.Text("BOCE 开放平台（申请 API Key）", nil, a.openURL(boceAPIURL)),
		menu.Separator(),
		menu.Text(aboutDialogTitle, nil, a.onAbout),
	)
}

// act 返回转发到前端的菜单处理器。
// 闭包在【调用时】才经 emitMenuAction 读 a.ctx —— 构造菜单时 startup 还没跑，a.ctx 是 nil。
func (a *App) act(name string) menu.Callback {
	return func(*menu.CallbackData) {
		a.emitMenuAction(name)
	}
}

func (a *App) openURL(url string) menu.Callback {
	return func(*menu.CallbackData) {
		if a.ctx == nil {
			return
		}
		wailsrt.BrowserOpenURL(a.ctx, url)
	}
}

// onAbout 显示版本信息。macOS 应用菜单里的「About」由 options.Mac.About 生成（见 main.go），
// 这里是 windows/linux 帮助菜单 + darwin 帮助菜单里的同名入口。
func (a *App) onAbout(*menu.CallbackData) {
	if a.ctx == nil {
		return
	}
	_, _ = wailsrt.MessageDialog(a.ctx, wailsrt.MessageDialogOptions{
		Type:    wailsrt.InfoDialog,
		Title:   aboutDialogTitle,
		Message: fmt.Sprintf("BOCE 检测工具\n版本 %s\n\n域名批量检测：DNS 污染 / QQ 拦截 / 微信拦截 / ICP 备案 / 备案黑名单 / 被墙。\nAPI Key 在 %s 申请。", Version, boceHomeURL),
	})
}

// onQuit —— 菜单「退出」：同步弹确认框，确认后再 Quit。
//
// 【约束】确认框绝不能放进 OnBeforeClose：macOS/Linux 在 goroutine 里异步调用它，
// 弹框时序不可控。这里同步调 MessageDialog 才是稳妥做法。
//
// 另注：runtime.Quit() 在三个平台上都会再走一遍 OnBeforeClose，所以确认放行后要先置
// quitConfirmed，否则检测中退出会被自己刚设的拦截再次挡下。
func (a *App) onQuit(*menu.CallbackData) {
	if a.ctx == nil {
		return
	}
	choice, err := wailsrt.MessageDialog(a.ctx, wailsrt.MessageDialogOptions{
		Type:          wailsrt.QuestionDialog,
		Title:         quitDialogTitle,
		Message:       "确定要退出 BOCE 检测工具吗？\n检测进行中退出会中断当前任务。",
		Buttons:       []string{quitConfirmOK, quitConfirmCancel},
		DefaultButton: quitConfirmCancel,
		CancelButton:  quitConfirmCancel,
	})
	if err != nil || choice != quitConfirmOK {
		return
	}
	quitConfirmed.Store(true)
	wailsrt.Quit(a.ctx)
}

func (a *App) onMinimize(*menu.CallbackData) {
	if a.ctx == nil {
		return
	}
	wailsrt.WindowMinimise(a.ctx)
}

func (a *App) onToggleMaximise(*menu.CallbackData) {
	if a.ctx == nil {
		return
	}
	wailsrt.WindowToggleMaximise(a.ctx)
}

func (a *App) onToggleFullscreen(*menu.CallbackData) {
	if a.ctx == nil {
		return
	}
	if wailsrt.WindowIsFullscreen(a.ctx) {
		wailsrt.WindowUnfullscreen(a.ctx)
		return
	}
	wailsrt.WindowFullscreen(a.ctx)
}

// onToggleAlwaysOnTop 用 MenuItem.Checked 作为新状态：
// Windows（windows/menu.go toggleCheckBox）和 Linux（linux/gtk.go）都会在调 Click 前
// 先翻转 Checked，这里直接采用即可。
func (a *App) onToggleAlwaysOnTop(data *menu.CallbackData) {
	if a.ctx == nil || data == nil || data.MenuItem == nil {
		return
	}
	wailsrt.WindowSetAlwaysOnTop(a.ctx, data.MenuItem.Checked)
}
