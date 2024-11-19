package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/multiformats/go-multiaddr"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

// HandleMessage 实现订阅处理程序
func (handler *SubscriptionHandler) HandleMessage(message *pubsub.Message) {
	fmt.Println("收到一条消息")
	var gossip Gossip
	err := json.Unmarshal(message.Data, &gossip)
	if err != nil {
		logrus.Errorf("[-] 解析消息失败: %v", err)
		return
	}

	handler.mu.Lock()
	handler.Gossips = append(handler.Gossips, gossip)
	handler.mu.Unlock()

	// logrus.Infof("已添加: %+v", gossip)
}

// showMenu 创建并返回菜单组件。
func handleMenu(app *tview.Application, output *tview.TextView) *tview.List {
	menu := tview.NewList().
		AddItem("连接", "[darkgray]连接到一个oracle节点", '1', func() {
			handleOption(app, "1", output)
		}).
		// 项目 2-5 已移除 - LLM 功能在 https://github.com/masa-finance/masa-oracle/pull/626 中被删除
		AddItem("Oracle 节点", "[darkgray]查看活动节点", '6', func() {
			handleOption(app, "6", output)
		})

	menu.AddItem("退出", "[darkgray]按下退出", 'q', func() {
		handleOption(app, "7", output)
	}).SetBorder(true).SetBorderColor(tcell.ColorGray)

	return menu
}

// handleOption 根据用户选择触发操作。
func handleOption(app *tview.Application, option string, output *tview.TextView) {
	switch option {
	case "1":
		modalFlex := tview.NewFlex().SetDirection(tview.FlexRow)
		modalFlex.SetBorderPadding(1, 1, 1, 1)

		var form *tview.Form

		// 创建一个新表单
		form = tview.NewForm().
			AddInputField("节点多地址", "", 60, nil, nil).
			AddButton("确定", func() {
				inputValue := form.GetFormItemByLabel("节点多地址").(*tview.InputField).GetText()
				appConfig.Address = inputValue

				if appConfig.Address == "" {
					output.SetText("未输入多地址。请输入Masa节点多地址并重试。")
				} else {
					output.SetText(fmt.Sprintf("连接到: %s", appConfig.Address))

					maddr, err := multiaddr.NewMultiaddr(appConfig.Address)
					if err != nil {
						logrus.Errorf("[-] %v", err)
					}

					// 创建一个libp2p主机以连接到Masa节点
					host, err := libp2p.New(libp2p.NoSecurity, libp2p.Transport(quic.NewTransport))
					if err != nil {
						logrus.Errorf("[-] %v", err)
					}

					// 从多地址中提取对等ID
					peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
					if err != nil {
						logrus.Errorf("[-] %v", err)
					}

					// 连接到对等节点
					if err := host.Connect(context.Background(), *peerInfo); err != nil {
						logrus.Errorf("[-] %v", err)
					}

					output.SetText(fmt.Sprintf("成功连接到节点: %s", appConfig.Address))
				}
				app.SetRoot(mainFlex, true) // 返回主视图
			}).
			AddButton("取消", func() {
				output.SetText("取消输入Masa节点多地址。")
				app.SetRoot(mainFlex, true) // 返回主视图
			})

		form.SetBorder(true).SetBorderColor(tcell.ColorBlue)

		modalFlex.AddItem(form, 0, 1, true)

		app.SetRoot(modalFlex, true).SetFocus(form)
	case "6":
		content := Splash()

		table := tview.NewTable().SetBorders(true).SetFixed(1, 0)

		// 设置标题
		table.SetCell(0, 0, tview.NewTableCell("地址").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))
		table.SetCell(0, 1, tview.NewTableCell("是否质押").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))
		table.SetCell(0, 2, tview.NewTableCell("是否验证者").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))

		// 设置每行的单元格值
		for i := 1; i <= 10; i++ {
			table.SetCell(i, 0, tview.NewTableCell("/ip4/127.0.0.1/udp/4001/quic-v1/p2p/16Uiu2HAmVRNDAZ6J1eHTV8twU6VaX8vqhe7VehPBNrCzDrHB9aQn"))
			table.SetCell(i, 1, tview.NewTableCell("false"))
			table.SetCell(i, 2, tview.NewTableCell("false"))
		}

		table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				mainFlex.Clear().
					AddItem(content, 0, 1, false).
					AddItem(handleMenu(app, output), 0, 1, true).
					AddItem(output, 0, 3, false)

				app.SetRoot(mainFlex, true) // 返回主视图
				return
			}
			if key == tcell.KeyEnter {
				table.SetSelectable(true, true)
			}
		}).SetSelectedFunc(func(row int, column int) {
			table.GetCell(row, column).SetTextColor(tcell.ColorRed)
			table.SetSelectable(false, false)
		})

		flex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(table, 0, 1, false)

		mainFlex.Clear().
			AddItem(content, 0, 1, false).
			AddItem(handleMenu(app, output), 0, 1, false).
			AddItem(flex, 0, 3, true)

		flex.SetBorder(true).SetBorderColor(tcell.ColorBlue).
			SetTitle(" Masa Oracle 节点，按 esc 返回菜单 ")

		app.SetRoot(mainFlex, true).SetFocus(table)
	case "7":
		modalFlex := tview.NewFlex().SetDirection(tview.FlexRow)
		modalFlex.SetBorderPadding(1, 1, 1, 1)

		modal := tview.NewModal().
			SetText("您确定要退出吗？").
			AddButtons([]string{"是", "否"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "是" {
					app.Stop()
				}
				app.SetRoot(mainFlex, true) // 返回主视图
			})

		modalFlex.AddItem(modal, 0, 1, true)

		app.SetRoot(modalFlex, true)
	default:
		break
	}
}

