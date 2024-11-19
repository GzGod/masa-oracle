package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/multiformats/go-multiaddr"

	"github.com/Gzgod/masa-oracle/internal/versioning"

	"github.com/sirupsen/logrus"

	"github.com/Gzgod/masa-oracle/node"
	"github.com/Gzgod/masa-oracle/pkg/api"
	"github.com/Gzgod/masa-oracle/pkg/config"
	"github.com/Gzgod/masa-oracle/pkg/db"
	"github.com/Gzgod/masa-oracle/pkg/staking"
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)
	logrus.Debug("日志级别设置为调试")

	if len(os.Args) > 1 && os.Args[1] == "--version" {
		logrus.Infof("Masa Oracle 节点版本: %s\nMasa Oracle 协议版本: %s", versioning.ApplicationVersion, versioning.ProtocolVersion)
		os.Exit(0)
	}

	cfg, err := config.GetConfig()
	if err != nil {
		logrus.Fatalf("[-] %v", err)
	}

	cfg.SetupLogging()
	cfg.LogConfig()

	// 创建一个可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())

	if cfg.Faucet {
		err := handleFaucet(cfg.RpcUrl, cfg.KeyManager.EcdsaPrivKey)
		if err != nil {
			logrus.Errorf("[-] %v", err)
			os.Exit(1)
		} else {
			logrus.Info("[+] 这个地址的水龙头事件已完成")
			os.Exit(0)
		}
	}

	if cfg.StakeAmount != "" {
		err := handleStaking(cfg.RpcUrl, cfg.KeyManager.EcdsaPrivKey, cfg.StakeAmount)
		if err != nil {
			logrus.Warningf("%v", err)
		} else {
			logrus.Info("[+] 这个地址的质押事件已完成")
			os.Exit(0)
		}
	}

	// 验证质押事件
	isStaked, err := staking.VerifyStakingEvent(cfg.RpcUrl, cfg.KeyManager.EthAddress)
	if err != nil {
		logrus.Error(err)
	}

	if !isStaked {
		logrus.Warn("没有找到这个地址的质押事件")
	}

	masaNodeOptions, workHandlerManager, pubKeySub := config.InitOptions(cfg)
	// 创建一个新的 OracleNode
	masaNode, err := node.NewOracleNode(ctx, masaNodeOptions...)

	if err != nil {
		logrus.Fatal(err)
	}

	if err = masaNode.Start(); err != nil {
		logrus.Fatal(err)
	}

	if cfg.TwitterScraper && cfg.DiscordScraper && cfg.WebScraper {
		logrus.Warn("[+] 节点被设置为所有类型的爬虫。这可能不是预期的行为。")
	}

	if cfg.AllowedPeer {
		cfg.AllowedPeerId = masaNode.Host.ID().String()
		cfg.AllowedPeerPublicKey = cfg.KeyManager.HexPubKey
		logrus.Infof("[+] 允许的对等节点 ID: %s 和公钥: %s", cfg.AllowedPeerId, cfg.AllowedPeerPublicKey)
	} else {
		logrus.Warn("[-] 这个节点没有设置为允许的对等节点")
	}

	// 初始化缓存解析器
	db.InitResolverCache(masaNode, cfg.KeyManager, cfg.AllowedPeerId, cfg.AllowedPeerPublicKey, cfg.Validator)

	// 收到 SIGINT 时取消上下文
	go handleSignals(cancel, masaNode, cfg)

	if cfg.APIEnabled {
		router := api.SetupRoutes(masaNode, workHandlerManager, pubKeySub)
		go func() {
			if err := router.Run(); err != nil {
				logrus.Fatal(err)
			}
		}()
		logrus.Info("API 服务器已启动")
	} else {
		logrus.Info("API 服务器已禁用")
	}

	// 获取节点的多地址和 IP 地址
	multiAddr := masaNode.GetMultiAddrs()                      // 获取多地址
	ipAddr, err := multiAddr.ValueForProtocol(multiaddr.P_IP4) // 获取 IP 地址
	if err != nil {
		logrus.Errorf("[-] 从 %v 获取节点 IP 地址时出错: %v", multiAddr, err)
	}
	// 显示包含多地址和 IP 地址的欢迎消息
	config.DisplayWelcomeMessage(multiAddr.String(), ipAddr, cfg.KeyManager.EthAddress, isStaked, cfg.Validator, cfg.TwitterScraper, cfg.TelegramScraper, cfg.DiscordScraper, cfg.WebScraper, versioning.ApplicationVersion, versioning.ProtocolVersion)

	<-ctx.Done()
}

func handleSignals(cancel context.CancelFunc, masaNode *node.OracleNode, cfg *config.AppConfig) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	nodeData := masaNode.NodeTracker.GetNodeData(masaNode.Host.ID().String())
	if nodeData != nil {
		nodeData.Left()
	}
	cancel()
	if cfg.TelegramStop != nil {
		if err := cfg.TelegramStop(); err != nil {
			logrus.Errorf("停止后台连接时出错: %v", err)
		}
	}
}
