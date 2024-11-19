package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/Gzgod/masa-oracle/internal/versioning"

	"github.com/rivo/tview"
)

// NewInputBox 返回一个新的输入框组件。
func NewInputBox() *InputBox {
	textView := tview.NewTextView().SetDynamicColors(true).SetRegions(true)
	return &InputBox{
		Box:      tview.NewBox().SetBorder(true).SetTitle("Input"),
		input:    make(chan rune),
		textView: textView,
	}
}

// InputHandler 返回一个处理 InputBox 键盘输入事件的函数。
// 它监听字符键的输入并将字符发送到 InputBox 的输入通道。
func (i *InputBox) InputHandler() func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			i.input <- event.Rune()
		}
		return event
	}
}

// Draw 在提供的屏幕上渲染 InputBox。
func (i *InputBox) Draw(screen tcell.Screen) {
	i.Box.DrawForSubclass(screen, i.Box)
	x, y, width, height := i.GetInnerRect()
	i.textView.SetRect(x, y, width, height)
	i.textView.Draw(screen)
}

// NewRadioButtons 返回一个新的单选按钮组件。
func NewRadioButtons(options []string, onSelect func(option string)) *RadioButtons {
	return &RadioButtons{
		Box:      tview.NewBox(),
		options:  options,
		onSelect: onSelect,
	}
}

// Draw 在屏幕上绘制这个组件。
func (r *RadioButtons) Draw(screen tcell.Screen) {
	r.Box.DrawForSubclass(screen, r)
	x, y, width, height := r.GetInnerRect()

	for index, option := range r.options {
		if index >= height {
			break
		}
		radioButton := "\u25ef" // 未选中。
		if index == r.currentOption {
			radioButton = "\u25c9" // 选中。
		}
		line := fmt.Sprintf(`%s[white]  %s`, radioButton, option)
		tview.Print(screen, line, x, y+index, width, tview.AlignLeft, tcell.ColorYellow)
	}
}

// InputHandler 返回这个组件的输入处理器。
func (r *RadioButtons) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			r.currentOption--
			if r.currentOption < 0 {
				r.currentOption = 0
			}
		case tcell.KeyDown:
			r.currentOption++
			if r.currentOption >= len(r.options) {
				r.currentOption = len(r.options) - 1
			}
		case tcell.KeyEnter:
			if r.onSelect != nil {
				r.onSelect(r.options[r.currentOption]) // 调用 onSelect 回调并传递选中的选项
			}
		}
	})
}

// MouseHandler 返回这个组件的鼠标处理器。
func (r *RadioButtons) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return r.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		x, y := event.Position()
		_, rectY, _, _ := r.GetInnerRect()
		if !r.InRect(x, y) {
			return false, nil
		}

		if action == tview.MouseLeftClick {
			setFocus(r)
			index := y - rectY
			if index >= 0 && index < len(r.options) {
				r.currentOption = index
				consumed = true
				if r.onSelect != nil {
					r.onSelect(r.options[r.currentOption]) // 调用回调并传递选中的选项
					// 关闭单选按钮视图的逻辑在这里
				}
			}
		}
		return
	})
}

const logo = ` 
  _____ _____    ___________   
 /     \\__  \  /  ___/\__  \  
|  Y Y  \/ __ \_\___ \  / __ \_
|__|_|  (____  /____  >(____  /
      \/     \/     \/      \/ 
`

const (
	subtitle   = `masa oracle client`
	navigation = `[yellow]使用键盘或鼠标导航`
)

var version string = fmt.Sprintf(`[green]应用版本: %s\n[green]协议版本: %s`, versioning.ApplicationVersion, versioning.ProtocolVersion)

// Splash 显示应用信息
func Splash() (content tview.Primitive) {
	lines := strings.Split(logo, "\n")
	logoWidth := 0
	logoHeight := len(lines)
	for _, line := range lines {
		if len(line) > logoWidth {
			logoWidth = len(line)
		}
	}
	logoBox := tview.NewTextView().
		SetTextColor(tcell.ColorGreen).
		SetDoneFunc(func(key tcell.Key) {
			// 无需处理
		})
	fmt.Fprint(logoBox, logo)

	frame := tview.NewFrame(tview.NewBox()).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText(subtitle, true, tview.AlignCenter, tcell.ColorWhite).
		AddText("", true, tview.AlignCenter, tcell.ColorWhite).
		AddText(navigation, true, tview.AlignCenter, tcell.ColorDarkMagenta).
		AddText(version, true, tview.AlignCenter, tcell.ColorDarkMagenta)

	// 创建一个 Flex 布局，居中显示 logo 和副标题。
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewBox(), 0, 7, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(logoBox, logoWidth, 1, false).
			AddItem(tview.NewBox(), 0, 1, false), logoHeight, 1, false).
		AddItem(frame, 0, 10, false)

	return flex
}
