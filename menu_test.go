package main

import (
	"runtime"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/menu"
)

// TestEmitMenuActionNilCtxSafe 守住菜单接线约定：
// 构造菜单时 startup 还没跑、a.ctx 仍是 nil，Click 闭包若直接解引用 ctx 会 panic。
func TestEmitMenuActionNilCtxSafe(t *testing.T) {
	app := &App{}
	if app.ctx != nil {
		t.Fatalf("新建 App 的 ctx = %v, want nil", app.ctx)
	}

	// 不应 panic
	app.emitMenuAction(menuActionStart)
	app.emitMenuAction(menuActionImportTXT)
	app.emitMenuAction(menuActionCloseBlocked)
}

// TestGetVersionDefault 版本占位：未用 -ldflags 注入时必须是 "dev"。
func TestGetVersionDefault(t *testing.T) {
	app := &App{}
	if got := app.GetVersion(); got != "dev" {
		t.Fatalf("GetVersion() = %q, want %q", got, "dev")
	}
	if got := app.GetVersion(); got != Version {
		t.Fatalf("GetVersion() = %q, want Version = %q", got, Version)
	}
}

// TestBuildMenuDarwinShape 断言菜单树结构。
//
// 通用部分（顶层项必须是 Role 或带 SubMenu 的子菜单）在所有平台都跑；
// darwin 专属的 3 个 Role 项断言在非 darwin 上 t.Skip —— Role 项在 Windows/Linux
// 会被静默丢弃，Windows 顶层还会多出一个点不开的空白标题菜单，两边的树本来就不一样。
func TestBuildMenuDarwinShape(t *testing.T) {
	app := NewApp()
	built := app.BuildMenu()
	if built == nil {
		t.Fatal("BuildMenu() = nil, want non-nil")
	}
	if len(built.Items) == 0 {
		t.Fatal("BuildMenu() 顶层为空")
	}

	// 通用断言：每个顶层项要么是 Role 项，要么是带 SubMenu 的子菜单。
	// 否则 Windows 的 processMenu 会无条件 AddSubMenu(空 Label)，多出一个空白标题菜单。
	roleCount := 0
	for index, item := range built.Items {
		if item == nil {
			t.Fatalf("顶层第 %d 项为 nil", index)
		}
		if item.Role != 0 {
			roleCount++
			continue
		}
		if item.SubMenu == nil {
			t.Errorf("顶层第 %d 项 %q 既不是 Role 项也没有子菜单", index, item.Label)
		}
	}

	if runtime.GOOS != "darwin" {
		if roleCount != 0 {
			t.Errorf("非 darwin 平台出现 %d 个 Role 项（会被静默丢弃并多出空白标题菜单）, want 0", roleCount)
		}
		t.Skip("darwin 专属 Role 项断言仅在 darwin 上执行")
	}

	// darwin：恰好 3 个 Role 项，顺序为 AppMenu / EditMenu / WindowMenu，中间夹着自定义子菜单。
	// 漏任何一个都会丢掉一整块系统能力（⌘Q / ⌘V / ⌘M）。
	wantTopLevel := []struct {
		isRole bool
		label  string
		role   menu.Role
	}{
		{isRole: true, role: menu.AppMenuRole},
		{isRole: false, label: "文件"},
		{isRole: true, role: menu.EditMenuRole},
		{isRole: false, label: "检测"},
		{isRole: true, role: menu.WindowMenuRole},
		{isRole: false, label: "帮助"},
	}

	if len(built.Items) != len(wantTopLevel) {
		t.Fatalf("darwin 顶层项数 = %d, want %d", len(built.Items), len(wantTopLevel))
	}
	for index, want := range wantTopLevel {
		got := built.Items[index]
		if want.isRole {
			if got.Role != want.role {
				t.Errorf("darwin 第 %d 项 Role = %d, want %d", index, got.Role, want.role)
			}
			if got.SubMenu != nil {
				t.Errorf("darwin 第 %d 个 Role 项不应带自定义子菜单（Role 子菜单是封闭的）", index)
			}
			continue
		}
		if got.Role != 0 {
			t.Errorf("darwin 第 %d 项 Role = %d, want 0（自定义子菜单）", index, got.Role)
		}
		if got.Label != want.label {
			t.Errorf("darwin 第 %d 项 Label = %q, want %q", index, got.Label, want.label)
		}
	}

	if roleCount != 3 {
		t.Errorf("darwin Role 项数 = %d, want 3", roleCount)
	}
}
