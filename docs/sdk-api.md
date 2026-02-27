# SDK API 设计文档

Go 版本小艺 Agent SDK 的对外 API 设计。

## 快速开始

```go
c := client.New(cfg)
c.OnMessage(handleMessage)
c.Connect(ctx)
defer c.Close()
```

## 核心接口

```go
// Client - SDK 客户端
type Client interface {
    // 连接管理
    Connect(ctx context.Context) error
    Close() error
    IsReady() bool
    
    // 消息发送
    Reply(ctx context.Context, taskID, sessionID, text string) error
    ReplyStream(ctx context.Context, taskID, sessionID, text string, isFinal, append bool) error
    SendStatus(ctx context.Context, taskID, sessionID, message, state string) error
    SendError(ctx context.Context, taskID, sessionID, code, message string) error
    
    // 事件注册
    OnMessage(handler MessageHandler)
    OnClear(handler func(sessionID string))
    OnCancel(handler func(sessionID, taskID string))
    OnError(handler func(serverID string, err error))
}

// Message - 接收到的消息
type Message interface {
    TaskID() string
    SessionID() string
    Text() string
    Parts() []Part
}

// Part - 消息部分
type Part interface {
    Kind() string
}

type TextPart interface {
    Part
    Text() string
}

type FilePart interface {
    Part
    Name() string
    MimeType() string
    URI() string
    Bytes() []byte
}

// MessageHandler - 消息处理器
type MessageHandler func(ctx context.Context, msg Message) error
```

## 配置

```go
type Config struct {
    AK              string
    SK              string
    AgentID         string
    WSUrl1          string        // 可选，默认小艺服务器1
    WSUrl2          string        // 可选，默认小艺服务器2
    EnableStreaming bool          // 默认 true
    SingleServer    bool          // 默认 false，只连接 server1
    ReconnectDelay  time.Duration // 默认 10s，重连基础延迟
}

func New(cfg *Config) Client
func DefaultConfig() *Config
```

## 错误

```go
type XiaoYiError struct {
    Code    string
    Message string
}

var (
    ErrNotConnected    = &XiaoYiError{Code: "NOT_CONNECTED"}
    ErrSessionNotFound = &XiaoYiError{Code: "SESSION_NOT_FOUND"}
    ErrConfigInvalid   = &XiaoYiError{Code: "CONFIG_INVALID"}
    ErrServerNotReady  = &XiaoYiError{Code: "SERVER_NOT_READY"}
)
```

## 示例

```go
func main() {
    cfg := &client.Config{
        AK:      os.Getenv("XIAOYI_AK"),
        SK:      os.Getenv("XIAOYI_SK"),
        AgentID: os.Getenv("XIAOYI_AGENT_ID"),
    }
    
    c := client.New(cfg)
    
    c.OnMessage(func(ctx context.Context, msg client.Message) error {
        // 处理文件
        for _, p := range msg.Parts() {
            if f, ok := p.(client.FilePart); ok {
                log.Printf("文件: %s", f.Name())
            }
        }
        // 回复
        return c.Reply(ctx, msg.TaskID(), msg.SessionID(), "收到")
    })
    
    c.OnCancel(func(sessionID, taskID string) {
        log.Printf("取消: %s", taskID)
    })
    
    if err := c.Connect(context.Background()); err != nil {
        log.Fatal(err)
    }
    defer c.Close()
    
    select {}
}
```
