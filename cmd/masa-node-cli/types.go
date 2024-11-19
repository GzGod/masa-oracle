package main

import (
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/rivo/tview"
)

type 应用配置 struct {
	地址         string
	模型           string
	推特用户     string
	推特密码 string
	推特2FA      string
}

var 应用配置实例 = 应用配置{}

var 主Flex *tview.Flex

type 八卦 struct {
	内容  string
	元数据 map[string]string
}

type 讲话请求 struct {
	文本          string `json:"text"`
	语音设置 struct {
		稳定性       float64 `json:"stability"`
		相似度提升 float64 `json:"similarity_boost"`
	} `json:"voice_settings"`
}

type 订阅处理器 struct {
	八卦数组     []八卦
	八卦主题 *pubsub.Topic
	互斥锁          sync.Mutex
}

type 单选按钮 struct {
	*tview.Box
	选项       []string
	当前选项 int
	选择回调      func(option string)
}

type 输入框 struct {
	*tview.Box
	输入    chan rune
	文本视图 *tview.TextView
}
