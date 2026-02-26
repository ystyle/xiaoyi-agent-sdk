# AGENTS.md - XiaoYi Agent SDK

Go 语言版本的小艺 Agent SDK，用于 PicoClaw 接入小艺的 A2A (Agent-to-Agent) 协议。

## 构建/测试命令

```bash
go build ./...                          # 构建项目
go test ./...                           # 运行所有测试
go test -v -run TestAuth ./pkg/auth     # 运行单个测试
go test -cover ./...                    # 测试覆盖率
go vet ./...                            # 代码检查
go fmt ./...                            # 格式化
go mod tidy                             # 整理依赖
```

## 项目结构

```
pkg/          # 公开 API (client, websocket, auth, types)
internal/     # 内部实现 (protocol, util)
examples/     # 示例代码
docs/         # 文档
  protocol/   # 协议文档
package/      # TypeScript 参考实现 (只读)
```

## 代码风格

### Imports (标准库 → 第三方 → 本项目)

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/gorilla/websocket"

    "github.com/ystyle/xiaoyi-agent-sdk/pkg/types"
)
```

### 命名

- 包名: 小写无下划线 (`websocket`, `auth`)
- 导出类型: PascalCase (`XiaoYiClient`, `A2AMessage`)
- 私有字段: camelCase (`wsConn`, `authHeader`)
- 接口: 动词或 `-er` 结尾 (`MessageHandler`)

### 错误处理

```go
type XiaoYiError struct {
    Code    string
    Message string
    Err     error
}

func (e *XiaoYiError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 错误包装
return fmt.Errorf("failed to connect: %w", err)
```

### Context (所有阻塞操作必须接受)

```go
func (c *Client) Connect(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    // ...
}
```

### 并发

```go
type Client struct {
    sendChan chan *Message
    done     chan struct{}
    wg       sync.WaitGroup
    mu       sync.RWMutex
}

func (c *Client) startWriter() {
    c.wg.Add(1)
    go func() {
        defer c.wg.Done()
        for {
            select {
            case msg := <-c.sendChan:
                c.writeMessage(msg)
            case <-c.done:
                return
            }
        }
    }()
}
```

## A2A 协议

详细协议文档见 [docs/protocol/a2a.md](docs/protocol/a2a.md)

## 关键实现要点

1. **双服务器**: 同时连接 server1 和 server2，会话绑定到特定服务器
2. **会话路由**: 使用 `map[sessionId]serverID` 维护映射
3. **WSS+IP**: Server 2 使用 IP 地址，需跳过 TLS 证书验证
4. **心跳**: 协议层 30s ping，应用层 20s heartbeat，90s 超时重连
5. **重连**: 指数退避，最大延迟 60s，最大重试 50 次
6. **线程安全**: Client 必须支持并发调用
7. **资源清理**: 断开时清理所有定时器和 goroutine

## 参考

- TypeScript 参考实现: `package/dist/`
- 协议文档: [docs/protocol/a2a.md](docs/protocol/a2a.md)
