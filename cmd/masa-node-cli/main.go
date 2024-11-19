package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gdamore/tcell/v2"
	"github.com/joho/godotenv"
	"github.com/rivo/tview"
)

func main() {
	var err error
	_, b, _, _ := runtime.Caller(0) // 获取当前文件的路径
	rootDir := filepath.Join(filepath.Dir(b), "../..") // 获取项目根目录
	if _, _ = os.Stat(rootDir + "/.env"); !os.IsNotExist(err) {
		_ = godotenv.Load() // 加载环境变量文件
	}

	app := tview.NewApplication() // 创建一个新的tview应用程序

	output := tview.NewTextView().
		SetDynamicColors(true). // 设置动态颜色
		SetText(" 欢迎使用 MASA Oracle 客户端 "). // 设置文本
		SetTextAlign(tview.AlignCenter) // 设置文本对齐方式

	content := Splash() // 调用Splash函数获取内容

	mainFlex = tview.NewFlex().SetDirection(tview.FlexColumn). // 创建一个新的Flex布局
		AddItem(content, 0, 1, false). // 添加内容
		AddItem(handleMenu(app, output), 0, 1, true). // 添加菜单
		AddItem(output, 0, 3, false) // 添加输出视图

	output.SetBorder(true).SetBorderColor(tcell.ColorBlue) // 设置边框及其颜色

	app.SetFocus(handleMenu(app, output)) // 设置焦点到菜单

	if err := app.SetRoot(mainFlex, true).EnableMouse(true).Run(); err != nil { // 设置根视图并运行应用程序
		log.Fatal(err) // 如果运行失败，记录错误并退出
	}
}
