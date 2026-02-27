package client

import (
	"context"
	"fmt"
	"time"

	"github.com/ystyle/xiaoyi-agent-sdk/internal/protocol"
	"github.com/ystyle/xiaoyi-agent-sdk/pkg/types"
	"github.com/ystyle/xiaoyi-agent-sdk/pkg/websocket"
)

type Client interface {
	Connect(ctx context.Context) error
	Close() error
	IsReady() bool

	Reply(ctx context.Context, taskID, sessionID, text string) error
	ReplyStream(ctx context.Context, taskID, sessionID, text string, isFinal, append bool) error
	SendStatus(ctx context.Context, taskID, sessionID, message, state string) error
	SendError(ctx context.Context, taskID, sessionID, code, message string) error
	Push(ctx context.Context, sessionID, text string) error

	OnMessage(handler MessageHandler)
	OnClear(handler func(sessionID string))
	OnCancel(handler func(sessionID, taskID string))
	OnError(handler func(serverID string, err error))
}

type MessageHandler func(ctx context.Context, msg types.Message) error

type client struct {
	config  *types.Config
	manager *websocket.Manager
}

func New(cfg *types.Config) Client {
	cfg.ApplyDefaults()
	return &client{
		config:  cfg,
		manager: websocket.NewManager(cfg),
	}
}

func (c *client) Connect(ctx context.Context) error {
	if err := c.config.Validate(); err != nil {
		return err
	}
	return c.manager.Connect(ctx)
}

func (c *client) Close() error {
	c.manager.Close()
	return nil
}

func (c *client) IsReady() bool {
	return c.manager.IsReady()
}

func (c *client) Reply(ctx context.Context, taskID, sessionID, text string) error {
	return c.ReplyStream(ctx, taskID, sessionID, text, true, false)
}

func (c *client) ReplyStream(ctx context.Context, taskID, sessionID, text string, isFinal, append bool) error {
	if !c.IsReady() {
		return types.ErrNotConnected
	}

	messageID := protocol.GenerateID()
	parts := []types.Part{types.NewTextPart(text)}
	resp := protocol.BuildArtifactResponse(messageID, taskID, parts, isFinal, append)
	return c.manager.SendResponse(taskID, sessionID, resp)
}

func (c *client) SendStatus(ctx context.Context, taskID, sessionID, message, state string) error {
	if !c.IsReady() {
		return types.ErrNotConnected
	}

	messageID := protocol.GenerateID()
	resp := protocol.BuildStatusResponse(messageID, taskID, message, state)
	return c.manager.SendResponse(taskID, sessionID, resp)
}

func (c *client) SendError(ctx context.Context, taskID, sessionID, code, message string) error {
	if !c.IsReady() {
		return types.ErrNotConnected
	}

	messageID := protocol.GenerateID()
	resp := protocol.BuildErrorResponse(messageID, code, message)
	return c.manager.SendResponse(taskID, sessionID, resp)
}

func (c *client) Push(ctx context.Context, sessionID, text string) error {
	if !c.IsReady() {
		return types.ErrNotConnected
	}

	messageID := protocol.GenerateID()
	taskID := fmt.Sprintf("push_%s", messageID)
	parts := []types.Part{types.NewTextPart(text)}
	resp := protocol.BuildPushResponse(messageID, taskID, text, parts)
	return c.manager.SendResponse(taskID, sessionID, resp)
}

func (c *client) OnMessage(handler MessageHandler) {
	c.manager.OnMessage(func(msg *types.A2ARequest) {
		if err := handler(context.Background(), msg); err != nil {
			fmt.Printf("[ERROR] Message handler error: %v\n", err)
		}
	})
}

func (c *client) OnClear(handler func(sessionID string)) {
	c.manager.OnClear(handler)
}

func (c *client) OnCancel(handler func(sessionID, taskID string)) {
	c.manager.OnCancel(handler)
}

func (c *client) OnError(handler func(serverID string, err error)) {
	c.manager.OnError(func(id types.ServerID, err error) {
		handler(string(id), err)
	})
}

func GenerateMessageID() string {
	return protocol.GenerateID()
}

func NewResponse(sessionID, agentID, text string) *types.A2AResponse {
	return &types.A2AResponse{
		SessionID: sessionID,
		MessageID: GenerateMessageID(),
		Timestamp: time.Now().UnixMilli(),
		AgentID:   agentID,
		Sender: types.Sender{
			ID:   agentID,
			Name: "Agent",
			Type: "agent",
		},
		Content: types.ResponseContent{
			Type: "text",
			Text: text,
		},
		Status: "success",
	}
}
