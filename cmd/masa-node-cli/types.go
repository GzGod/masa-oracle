package main

import (
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/rivo/tview"
)

// 应用程序配置结构体
type AppConfig struct {
	Address         string // 地址
	Model           string // 模型
	TwitterUser     string // 推特用户名
	TwitterPassword string // 推特密码
	Twitter2FA      string // 推特两步验证
}

var appConfig = AppConfig{} // 初始化应用程序配置

var mainFlex *tview.Flex // 主布局

// Gossip 结构体，表示一个消息
type Gossip struct {
	Content  string            // 消息内容
	Metadata map[string]string // 消息元数据
}

// SpeakRequest 结构体，表示一个说话请求
type SpeakRequest struct {
	Text          string `json:"text"` // 文本内容
	VoiceSettings struct {
		Stability       float64 `json:"stability"`        // 稳定性
		SimilarityBoost float64 `json:"similarity_boost"` // 相似度提升
	} `json:"voice_settings"`
}

// SubscriptionHandler 结构体，处理订阅
type SubscriptionHandler struct {
	Gossips     []Gossip      // 消息列表
	GossipTopic *pubsub.Topic // 消息主题
	mu          sync.Mutex    // 互斥锁
}

// RadioButtons 结构体，表示一组单选按钮
type RadioButtons struct {
	*tview.Box
	options       []string               // 选项列表
	currentOption int                    // 当前选项索引
	onSelect      func(option string)    // 选项选择回调函数
}

// InputBox 结构体，表示一个输入框
type InputBox struct {
	*tview.Box
	input    chan rune         // 输入通道
	textView *tview.TextView   // 文本视图
}
