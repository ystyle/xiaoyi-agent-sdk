package main

import (
	"context"
	"fmt"
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
	godotenv.Load()

	cfg := &types.Config{
		AK:      os.Getenv("XIAOYI_AK"),
		SK:      os.Getenv("XIAOYI_SK"),
		AgentID: os.Getenv("XIAOYI_AGENT_ID"),
	}

	if cfg.AK == "" || cfg.SK == "" || cfg.AgentID == "" {
		fmt.Fprintln(os.Stderr, "请设置环境变量: XIAOYI_AK, XIAOYI_SK, XIAOYI_AGENT_ID")
		os.Exit(1)
	}

	c := client.New(cfg)

	var (
		firstMessageOnce sync.Once
		sessionTaskIDs   = make(map[string]string)
		sessionTaskIDsMu sync.Mutex
	)

	c.OnMessage(func(ctx context.Context, msg types.Message) error {
		sessionID := msg.SessionID()
		taskID := msg.TaskID()
		text := strings.TrimSpace(msg.Text())

		fmt.Printf("\n[MSG] Session: %s\n", sessionID)
		fmt.Printf("[MSG] TaskID: %s\n", taskID)
		fmt.Printf("[MSG] Text: %s\n", text)

		for _, p := range msg.Parts() {
			if f, ok := p.(*types.FilePart); ok {
				fmt.Printf("[MSG] File: %s (%s) URI: %s\n", f.Name(), f.MimeType(), f.URI())
			}
		}

		sessionTaskIDsMu.Lock()
		sessionTaskIDs[sessionID] = taskID
		sessionTaskIDsMu.Unlock()

		firstMessageOnce.Do(func() {
			fmt.Println("\n[FIRST] 开始每30秒推送测试消息...")
			go startPushLoop(c, &sessionTaskIDs, &sessionTaskIDsMu)
		})

		if strings.HasPrefix(text, "/long") {
			delay := time.Duration(rand.Intn(30)+1) * time.Second
			fmt.Printf("[LONG] 将在 %v 后响应...\n", delay)

			go func() {
				time.Sleep(delay)
				reply := fmt.Sprintf("长任务完成 (延迟 %v)", delay)
				if err := c.Reply(context.Background(), taskID, sessionID, reply); err != nil {
					fmt.Printf("[LONG] 发送失败: %v\n", err)
				} else {
					fmt.Printf("[LONG] 已响应\n")
				}
			}()
			return c.SendStatus(ctx, taskID, sessionID, fmt.Sprintf("处理中，预计 %v 后完成", delay))
		}

		reply := fmt.Sprintf("Echo: %s", text)
		return c.Reply(ctx, taskID, sessionID, reply)
	})

	c.OnClear(func(sessionID string) {
		fmt.Printf("\n[CLEAR] Session: %s\n", sessionID)
		sessionTaskIDsMu.Lock()
		delete(sessionTaskIDs, sessionID)
		sessionTaskIDsMu.Unlock()
	})

	c.OnCancel(func(sessionID, taskID string) {
		fmt.Printf("\n[CANCEL] Session: %s, Task: %s\n", sessionID, taskID)
	})

	c.OnError(func(serverID string, err error) {
		fmt.Printf("[ERROR] Server: %s, Error: %v\n", serverID, err)
	})

	fmt.Println("Connecting...")
	if err := c.Connect(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Connect failed: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	fmt.Println("Connected!")
	fmt.Println("命令:")
	fmt.Println("  任意消息 -> Echo 回复")
	fmt.Println("  /long   -> 随机延迟 1-30s 后回复")
	fmt.Println("\nWaiting for messages...")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
}

func startPushLoop(c client.Client, taskIDs *map[string]string, mu *sync.Mutex) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	counter := 1
	for range ticker.C {
		mu.Lock()
		sessCopy := make(map[string]string)
		for k, v := range *taskIDs {
			sessCopy[k] = v
		}
		mu.Unlock()

		if len(sessCopy) == 0 {
			continue
		}

		msg := fmt.Sprintf("[PUSH] 测试推送 #%d - %s", counter, time.Now().Format("15:04:05"))
		fmt.Printf("\n%s\n", msg)

		for sessionID := range sessCopy {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := c.Push(ctx, sessionID, msg); err != nil {
				fmt.Printf("[PUSH] 发送失败 Session=%s: %v\n", sessionID, err)
			} else {
				fmt.Printf("[PUSH] 已发送 Session=%s\n", sessionID)
			}
			cancel()
		}
		counter++
	}
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
