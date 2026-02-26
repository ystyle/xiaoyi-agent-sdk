# A2A 协议文档

小艺 Agent-to-Agent 协议的 WebSocket 通信规范。

## WebSocket URL

| 服务器 | URL | 说明 |
|--------|-----|------|
| Server 1 | `wss://hag.cloud.huawei.com/openclaw/v1/ws/link` | 主服务器 |
| Server 2 | `wss://116.63.174.231/openclaw/v1/ws/link` | 备用服务器 (IP 地址，需跳过 TLS 验证) |

## 认证机制

### AK/SK 签名

使用 HMAC-SHA256 生成签名：

```
signature = Base64(HMAC-SHA256(SK, timestamp))
```

其中 `timestamp` 为当前时间戳（毫秒）的字符串形式。

### WebSocket 连接头

```
x-access-key: Access Key
x-sign:       HMAC-SHA256 签名 (Base64 编码)
x-ts:         当前时间戳 (毫秒)
x-agent-id:   Agent 标识符
```

## 消息类型

### 1. 初始化消息 (客户端 → 服务端)

连接建立后立即发送：

```json
{
    "msgType": "clawd_bot_init",
    "agentId": "your-agent-id"
}
```

### 2. 心跳消息 (客户端 → 服务端)

每 20 秒发送一次：

```json
{
    "msgType": "heartbeat",
    "agentId": "your-agent-id"
}
```

### 3. A2A 请求 (服务端 → 客户端)

```json
{
    "jsonrpc": "2.0",
    "id": "request-id",
    "method": "message/stream",
    "agentId": "your-agent-id",
    "deviceId": "device-id",
    "conversationId": "conversation-id",
    "sessionId": "session-id",
    "params": {
        "id": "task-id",
        "sessionId": "session-id",
        "agentLoginSessionId": "login-session-id",
        "message": {
            "kind": "message-kind",
            "messageId": "message-id",
            "role": "user",
            "parts": [
                {
                    "kind": "text",
                    "text": "用户消息内容"
                }
            ]
        }
    }
}
```

### 4. Agent 响应 (客户端 → 服务端)

```json
{
    "msgType": "agent_response",
    "agentId": "your-agent-id",
    "sessionId": "session-id",
    "taskId": "task-id",
    "msgDetail": "{...JSON-RPC response...}"
}
```

### 5. 清除上下文 (服务端 → 客户端)

```json
{
    "jsonrpc": "2.0",
    "id": "request-id",
    "method": "clearContext",
    "agentId": "your-agent-id",
    "sessionId": "session-id"
}
```

响应：

```json
{
    "jsonrpc": "2.0",
    "id": "request-id",
    "result": {
        "status": {
            "state": "cleared"
        }
    }
}
```

### 6. 任务取消 (服务端 → 客户端)

```json
{
    "jsonrpc": "2.0",
    "id": "request-id",
    "method": "tasks/cancel",
    "agentId": "your-agent-id",
    "sessionId": "session-id",
    "taskId": "task-id"
}
```

响应：

```json
{
    "jsonrpc": "2.0",
    "id": "request-id",
    "result": {
        "id": "task-id",
        "status": {
            "state": "canceled"
        }
    }
}
```

## 消息 Parts 类型

### 文本消息

```json
{
    "kind": "text",
    "text": "消息文本内容"
}
```

### 文件消息

```json
{
    "kind": "file",
    "file": {
        "name": "filename.pdf",
        "mimeType": "application/pdf",
        "bytes": "base64-encoded-content",
        "uri": "https://example.com/file.pdf"
    }
}
```

### 数据消息

```json
{
    "kind": "data",
    "data": { ... }
}
```

## JSON-RPC 响应格式

### Artifact Update (流式/最终响应)

```json
{
    "jsonrpc": "2.0",
    "id": "message-id",
    "result": {
        "taskId": "task-id",
        "kind": "artifact-update",
        "append": false,
        "lastChunk": true,
        "final": true,
        "artifact": {
            "artifactId": "artifact-id",
            "parts": [
                {
                    "kind": "text",
                    "text": "响应内容"
                }
            ]
        }
    }
}
```

**字段说明**：
- `append`: 是否追加到之前的消息（流式响应时为 true）
- `lastChunk`: 是否为最后一个分片
- `final`: 是否为最终响应（true 表示对话结束）

### Status Update (状态更新)

```json
{
    "jsonrpc": "2.0",
    "id": "message-id",
    "result": {
        "taskId": "task-id",
        "kind": "status-update",
        "final": false,
        "status": {
            "message": {
                "role": "agent",
                "parts": [
                    {
                        "kind": "text",
                        "text": "正在处理..."
                    }
                ]
            },
            "state": "working"
        }
    }
}
```

### 错误响应

```json
{
    "jsonrpc": "2.0",
    "id": "message-id",
    "error": {
        "code": "ERROR_CODE",
        "message": "错误描述"
    }
}
```

## 任务状态

| 状态 | 说明 |
|------|------|
| `submitted` | 已提交 |
| `working` | 处理中 |
| `input-required` | 需要用户输入 |
| `completed` | 已完成 |
| `canceled` | 已取消 |
| `failed` | 失败 |
| `unknown` | 未知 |

## 心跳机制

| 类型 | 间隔 | 超时 |
|------|------|------|
| 协议层 (WebSocket Ping) | 30 秒 | 90 秒 |
| 应用层 (Heartbeat Message) | 20 秒 | - |

## 重连策略

- 指数退避：初始 2 秒，每次翻倍
- 最大延迟：60 秒
- 最大重试：50 次
- **稳定性检测**：连接稳定 10 秒后重置重连计数器

## 会话路由

每个会话绑定到特定的服务器：

1. 收到消息时，记录 `sessionId → serverId` 映射
2. 发送响应时，根据 `sessionId` 查找对应的服务器
3. 清除会话时，删除映射关系

## 流式响应

支持增量发送消息：

1. 首次响应：`append=false`, `final=false`
2. 增量更新：`append=true`, `final=false`
3. 最终响应：`append=false`, `final=true`

## 媒体文件处理

1. 收到文件消息后，从 `file.uri` 下载文件
2. 文本类型文件（txt, md, json 等）可提取内容
3. 二进制文件保存到本地磁盘

## 参考文档

- [华为消息流文档](https://developer.huawei.com/consumer/cn/doc/service/message-stream-0000002505761434)
- [华为推送消息文档](https://developer.huawei.com/consumer/cn/doc/service/pushmessage-0000002505761436)
