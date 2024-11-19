package main

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"

	"github.com/masa-finance/masa-oracle/pkg/staking"
)

func handleStaking(rpcUrl string, privateKey *ecdsa.PrivateKey, stakeAmount string) error {
	// 质押逻辑
	// 将质押金额转换为最小单位，假设有18个小数位
	amountBigInt, ok := new(big.Int).SetString(stakeAmount, 10)
	if !ok {
		logrus.Fatal("无效的质押金额")
	}
	amountInSmallestUnit := new(big.Int).Mul(amountBigInt, big.NewInt(1e18))

	stakingClient, err := staking.NewClient(rpcUrl, privateKey)
	if err != nil {
		return err
	}
	// 启动和停止带有消息的旋转器的函数
	startSpinner := func(msg string, txHashChan <-chan string, done chan bool) {
		spinner := []string{"|", "/", "-", "\\"}
		i := 0
		var txHash string
		for {
			select {
			case txHash = <-txHashChan: // 接收交易哈希
				// 不在这里打印任何内容，只更新 txHash 变量
			case <-done:
				fmt.Printf("\r%s\n", msg) // 完成时打印最终消息
				if txHash != "" {
					fmt.Println(txHash) // 在新行上打印交易哈希
				}
				return
			default:
				// 使用回车符 `\r` 在同一行上覆盖旋转动画
				// 从打印语句中删除换行符 `\n`
				if txHash != "" {
					fmt.Printf("\r%s %s - %s", spinner[i], msg, txHash)
				} else {
					fmt.Printf("\r%s %s", spinner[i], msg)
				}
				i = (i + 1) % len(spinner)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	// 批准质押合约代表用户花费代币
	var approveTxHash string
	done := make(chan bool)
	txHashChan := make(chan string, 1) // 缓冲为1以防止阻塞
	go startSpinner("正在批准质押合约花费代币...", txHashChan, done)
	approveTxHash, err = stakingClient.Approve(amountInSmallestUnit)
	if err != nil {
		logrus.Error("[-] 批准质押代币失败:", err)
		return err
	}
	txHashChan <- approveTxHash // 将交易哈希发送到旋转器
	done <- true                // 停止旋转器
	color.Green("批准交易哈希: %s", approveTxHash)

	// 批准后质押代币
	var stakeTxHash string
	done = make(chan bool)
	txHashChan = make(chan string, 1) // 缓冲为1以防止阻塞
	go startSpinner("正在质押代币...", txHashChan, done)
	stakeTxHash, err = stakingClient.Stake(amountInSmallestUnit)
	if err != nil {
		logrus.Error("[-] 质押代币失败:", err)
		return err
	}
	txHashChan <- stakeTxHash // 将交易哈希发送到旋转器
	done <- true              // 停止旋转器
	color.Green("质押交易哈希: %s", stakeTxHash)

	return nil
}

func handleFaucet(rpcUrl string, privateKey *ecdsa.PrivateKey) error {
	faucetClient, err := staking.NewClient(rpcUrl, privateKey)
	if err != nil {
		logrus.Error("[-] 创建质押客户端失败:", err)
		return err
	}

	startSpinner := func(msg string, txHashChan <-chan string, done chan bool) {
		spinner := []string{"|", "/", "-", "\\"}
		i := 0
		var txHash string
		for {
			select {
			case txHash = <-txHashChan: // 接收交易哈希
				// 不在这里打印任何内容，只更新 txHash 变量
			case <-done:
				fmt.Printf("\r%s\n", msg) // 完成时打印最终消息
				if txHash != "" {
					fmt.Println(txHash) // 在新行上打印交易哈希
				}
				return
			default:
				// 使用回车符 `\r` 在同一行上覆盖旋转动画
				// 从打印语句中删除换行符 `\n`
				if txHash != "" {
					fmt.Printf("\r%s %s - %s", spinner[i], msg, txHash)
				} else {
					fmt.Printf("\r%s %s", spinner[i], msg)
				}
				i = (i + 1) % len(spinner)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	// 运行水龙头
	var faucetTxHash string
	done := make(chan bool)
	txHashChan := make(chan string, 1) // 缓冲为1以防止阻塞
	go startSpinner("正在请求水龙头代币...", txHashChan, done)
	faucetTxHash, err = faucetClient.RunFaucet()
	if err != nil {
		logrus.Error("[-] 请求水龙头代币失败:", err)
		return err
	}
	txHashChan <- faucetTxHash // 将交易哈希发送到旋转器
	done <- true               // 停止旋转器
	color.Green("[-] 水龙头交易哈希: %s", faucetTxHash)

	return nil
}
