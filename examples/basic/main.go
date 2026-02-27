package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/ystyle/xiaoyi-agent-sdk/pkg/client"
	"github.com/ystyle/xiaoyi-agent-sdk/pkg/types"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	godotenv.Load()

	cfg := &types.Config{
		AK:           os.Getenv("XIAOYI_AK"),
		SK:           os.Getenv("XIAOYI_SK"),
		AgentID:      os.Getenv("XIAOYI_AGENT_ID"),
		SingleServer: true,
	}

	if cfg.AK == "" || cfg.SK == "" || cfg.AgentID == "" {
		slog.Error("请设置环境变量: XIAOYI_AK, XIAOYI_SK, XIAOYI_AGENT_ID")
		os.Exit(1)
	}

	c := client.New(cfg)

	var (
		sessionTaskIDs   = make(map[string]string)
		sessionTaskIDsMu sync.Mutex
	)

	c.OnMessage(func(ctx context.Context, msg types.Message) error {
		sessionID := msg.SessionID()
		taskID := msg.TaskID()
		text := strings.TrimSpace(msg.Text())

		slog.Info("收到消息", "session", sessionID, "task", taskID, "text", text)

		for _, p := range msg.Parts() {
			if f, ok := p.(*types.FilePart); ok {
				slog.Info("收到文件", "name", f.Name(), "mime", f.MimeType(), "uri", f.URI())
			}
		}

		sessionTaskIDsMu.Lock()
		sessionTaskIDs[sessionID] = taskID
		sessionTaskIDsMu.Unlock()

		if strings.HasPrefix(text, "/long") {
			delay := time.Duration(rand.Intn(30)+1) * time.Second
			slog.Info("长任务开始", "delay", delay)

			go func() {
				time.Sleep(delay)
				reply := fmt.Sprintf("长任务完成 (延迟 %v)", delay)
				if err := c.ReplyStream(context.Background(), taskID, sessionID, reply, false, false); err != nil {
					slog.Error("长任务发送失败", "error", err)
					return
				}
				if err := c.SendStatus(context.Background(), taskID, sessionID, "已完成", "completed"); err != nil {
					slog.Error("发送完成状态失败", "error", err)
				}
				slog.Info("长任务已响应", "delay", delay)
			}()
			return c.SendStatus(ctx, taskID, sessionID, fmt.Sprintf("处理中，预计 %v 后完成", delay), "working")
		}

		reply := fmt.Sprintf("Echo: %s", text)
		return c.Reply(ctx, taskID, sessionID, reply)
	})

	c.OnClear(func(sessionID string) {
		slog.Info("清理会话", "session", sessionID)
		sessionTaskIDsMu.Lock()
		delete(sessionTaskIDs, sessionID)
		sessionTaskIDsMu.Unlock()
	})

	c.OnCancel(func(sessionID, taskID string) {
		slog.Info("取消任务", "session", sessionID, "task", taskID)
	})

	c.OnError(func(serverID string, err error) {
		slog.Error("服务器错误", "server", serverID, "error", err)
	})

	slog.Info("连接中...")
	if err := c.Connect(context.Background()); err != nil {
		slog.Error("连接失败", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	slog.Info("已连接", "commands", "任意消息->Echo回复, /long->随机延迟回复")
	slog.Info("等待消息...")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("关闭中...")
}
