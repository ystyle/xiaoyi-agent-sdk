# XiaoYi Agent SDK

Go 语言版本的小艺 Agent SDK，用于 **OpenClaw 智能体**接入华为小艺的 A2A (Agent-to-Agent) 协议。

> 本 SDK 为 OpenClaw 类型智能体设计，用于与小艺进行 Agent-to-Agent 通信。

## 特性

- WebSocket 长连接，支持双服务器
- AK/SK + HMAC-SHA256 认证
- 自动重连（指数退避）
- 心跳机制（协议层 + 应用层）
- 流式响应、状态更新
- 线程安全

## 安装

```bash
go get github.com/ystyle/xiaoyi-agent-sdk
```

## 快速开始

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"

    "github.com/ystyle/xiaoyi-agent-sdk/pkg/client"
    "github.com/ystyle/xiaoyi-agent-sdk/pkg/types"
)

func main() {
    cfg := &types.Config{
        AK:           os.Getenv("XIAOYI_AK"),
        SK:           os.Getenv("XIAOYI_SK"),
        AgentID:      os.Getenv("XIAOYI_AGENT_ID"),
        SingleServer: true,
    }

    c := client.New(cfg)

    c.OnMessage(func(ctx context.Context, msg types.Message) error {
        slog.Info("收到消息", "text", msg.Text())
        return c.Reply(ctx, msg.TaskID(), msg.SessionID(), "收到: "+msg.Text())
    })

    c.OnClear(func(sessionID string) {
        slog.Info("清理会话", "session", sessionID)
    })

    c.OnCancel(func(sessionID, taskID string) {
        slog.Info("取消任务", "task", taskID)
    })

    if err := c.Connect(context.Background()); err != nil {
        slog.Error("连接失败", "error", err)
        os.Exit(1)
    }
    defer c.Close()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
}
```

## 配置

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `AK` | string | Access Key | 必填 |
| `SK` | string | Secret Key | 必填 |
| `AgentID` | string | Agent 标识 | 必填 |
| `WSUrl1` | string | 服务器1 URL | 小艺服务器 |
| `WSUrl2` | string | 服务器2 URL | 小艺备用服务器 |
| `SingleServer` | bool | 只连接单个服务器 | false |
| `ReconnectDelay` | Duration | 重连基础延迟 | 10s |

## API

### 消息处理

```go
// 回复文本消息
c.Reply(ctx, taskID, sessionID, "回复内容")

// 发送状态更新
c.SendStatus(ctx, taskID, sessionID, "处理中...", "working")
c.SendStatus(ctx, taskID, sessionID, "已完成", "completed")

// 流式回复
c.ReplyStream(ctx, taskID, sessionID, "部分内容", false, true) // append=true

// 发送错误
c.SendError(ctx, taskID, sessionID, "ERROR_CODE", "错误描述")
```

### 事件注册

```go
c.OnMessage(func(ctx context.Context, msg types.Message) error {
    // msg.TaskID()    - 任务ID
    // msg.SessionID() - 会话ID
    // msg.Text()      - 文本内容
    // msg.Parts()     - 所有部分
    return nil
})

c.OnClear(func(sessionID string) {
    // 会话被清理
})

c.OnCancel(func(sessionID, taskID string) {
    // 任务被取消
})

c.OnError(func(serverID string, err error) {
    // 连接错误
})
```

### 消息类型

```go
// 获取消息部分
for _, p := range msg.Parts() {
    switch v := p.(type) {
    case *types.TextPart:
        fmt.Println(v.Text())
    case *types.FilePart:
        fmt.Println(v.Name(), v.MimeType(), v.URI())
    case *types.DataPart:
        fmt.Println(v.Data())
    }
}
```

## 心跳机制

| 类型 | 间隔 | 说明 |
|------|------|------|
| 协议层 Ping/Pong | 30s | WebSocket 原生心跳 |
| 应用层 Heartbeat | 20s | 业务心跳消息 |
| 超时断开 | 90s | 无响应则重连 |

## 重连策略

- 指数退避：10s → 20s → 40s → 60s (max)
- 最大重试：50 次
- 稳定检测：连接 10s 后重置计数器

## 示例

运行示例：

```bash
# 设置环境变量
export XIAOYI_AK=your-ak
export XIAOYI_SK=your-sk
export XIAOYI_AGENT_ID=your-agent-id

# 运行
go run examples/basic/main.go
```

## 文档

- [协议文档](docs/protocol/a2a.md)
- [API 设计](docs/sdk-api.md)

## License

MIT
