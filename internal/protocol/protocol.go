package protocol

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/ystyle/xiaoyi-agent-sdk/pkg/types"
)

func GenerateID() string {
	return fmt.Sprintf("%d_%s", time.Now().UnixMilli(), randomString(9))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func BuildInitMessage(agentID string) *types.OutboundMessage {
	return &types.OutboundMessage{
		MsgType: "clawd_bot_init",
		AgentID: agentID,
	}
}

func BuildHeartbeatMessage(agentID string) *types.OutboundMessage {
	return &types.OutboundMessage{
		MsgType: "heartbeat",
		AgentID: agentID,
	}
}

func BuildResponseMessage(agentID, sessionID, taskID string, jsonRpcResponse *types.JsonRpcResponse) *types.OutboundMessage {
	detail, _ := json.Marshal(jsonRpcResponse)
	return &types.OutboundMessage{
		MsgType:   "agent_response",
		AgentID:   agentID,
		SessionID: sessionID,
		TaskID:    taskID,
		MsgDetail: string(detail),
	}
}

func BuildArtifactResponse(messageID, taskID string, parts []types.Part, isFinal, append bool) *types.JsonRpcResponse {
	return &types.JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      messageID,
		Result: &types.ArtifactUpdate{
			TaskID:    taskID,
			Kind:      "artifact-update",
			Append:    append,
			LastChunk: isFinal,
			Final:     isFinal,
			Artifact: types.ArtifactPayload{
				ArtifactID: GenerateID(),
				Parts:      parts,
			},
		},
	}
}

func BuildStatusResponse(messageID, taskID, message string) *types.JsonRpcResponse {
	return &types.JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      messageID,
		Result: &types.StatusUpdate{
			TaskID: taskID,
			Kind:   "status-update",
			Final:  false,
			Status: types.StatusPayload{
				Message: types.MessageBody{
					Role: "agent",
					Parts: []types.Part{
						types.NewTextPart(message),
					},
				},
				State: "working",
			},
		},
	}
}

func BuildErrorResponse(messageID, code, message string) *types.JsonRpcResponse {
	return &types.JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      messageID,
		Error: &types.JsonRpcError{
			Code:    code,
			Message: message,
		},
	}
}

func BuildClearContextResponse(requestID string, success bool) *types.JsonRpcResponse {
	state := "cleared"
	if !success {
		state = "failed"
	}
	return &types.JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      requestID,
		Result: &types.ClearContextResult{
			Status: struct {
				State string `json:"state"`
			}{State: state},
		},
	}
}

func BuildTasksCancelResponse(requestID string, success bool) *types.JsonRpcResponse {
	state := "canceled"
	if !success {
		state = "failed"
	}
	return &types.JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      requestID,
		Result: &types.TasksCancelResult{
			ID: requestID,
			Status: struct {
				State string `json:"state"`
			}{State: state},
		},
	}
}

func BuildPushResponse(messageID, taskID, pushText string, parts []types.Part) *types.JsonRpcResponse {
	return &types.JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      messageID,
		Result: &types.PushUpdate{
			ID:       taskID,
			PushID:   GenerateID(),
			PushText: pushText,
			Kind:     "task",
			Artifacts: []types.PushArtifact{
				{
					ArtifactID: GenerateID(),
					Parts:      parts,
				},
			},
			Status: types.PushStatus{
				State: "completed",
			},
		},
	}
}

func ParseA2ARequest(data []byte) (*types.A2ARequest, error) {
	var raw struct {
		JSONRPC        string          `json:"jsonrpc"`
		ID             string          `json:"id"`
		Method         string          `json:"method"`
		AgentID        string          `json:"agentId"`
		DeviceID       string          `json:"deviceId,omitempty"`
		ConversationID string          `json:"conversationId,omitempty"`
		SessionID      string          `json:"sessionId,omitempty"`
		Params         json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	params, err := parseRequestParams(raw.Params)
	if err != nil {
		return nil, err
	}

	return &types.A2ARequest{
		JSONRPC:        raw.JSONRPC,
		ID:             raw.ID,
		Method:         raw.Method,
		AgentID:        raw.AgentID,
		DeviceID:       raw.DeviceID,
		ConversationID: raw.ConversationID,
		SessionIDField: raw.SessionID,
		Params:         *params,
	}, nil
}

func parseRequestParams(data json.RawMessage) (*types.RequestParams, error) {
	var raw struct {
		ID                  string          `json:"id"`
		SessionID           string          `json:"sessionId,omitempty"`
		AgentLoginSessionID string          `json:"agentLoginSessionId,omitempty"`
		Message             json.RawMessage `json:"message"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	message, err := parseMessageBody(raw.Message)
	if err != nil {
		return nil, err
	}

	return &types.RequestParams{
		ID:                  raw.ID,
		SessionIDField:      raw.SessionID,
		AgentLoginSessionID: raw.AgentLoginSessionID,
		Message:             *message,
	}, nil
}

func parseMessageBody(data json.RawMessage) (*types.MessageBody, error) {
	var raw struct {
		Kind      string          `json:"kind,omitempty"`
		MessageID string          `json:"messageId,omitempty"`
		Role      string          `json:"role"`
		Parts     json.RawMessage `json:"parts"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	parts, err := parseParts(raw.Parts)
	if err != nil {
		return nil, err
	}

	return &types.MessageBody{
		Kind:      raw.Kind,
		MessageID: raw.MessageID,
		Role:      raw.Role,
		Parts:     parts,
	}, nil
}

func parseParts(data json.RawMessage) ([]types.Part, error) {
	var rawParts []json.RawMessage
	if err := json.Unmarshal(data, &rawParts); err != nil {
		return nil, err
	}

	parts := make([]types.Part, 0, len(rawParts))
	for _, rp := range rawParts {
		part, err := parsePart(rp)
		if err != nil {
			continue
		}
		parts = append(parts, part)
	}
	return parts, nil
}

func parsePart(data json.RawMessage) (types.Part, error) {
	var kind struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &kind); err != nil {
		return nil, err
	}

	switch kind.Kind {
	case "text":
		var tp struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(data, &tp); err != nil {
			return nil, err
		}
		return types.NewTextPart(tp.Text), nil
	case "file":
		var fp struct {
			File types.File `json:"file"`
		}
		if err := json.Unmarshal(data, &fp); err != nil {
			return nil, err
		}
		return types.NewFilePart(fp.File.Name, fp.File.MimeType, fp.File.URI, fp.File.Bytes), nil
	case "data":
		var dp struct {
			Data any `json:"data"`
		}
		if err := json.Unmarshal(data, &dp); err != nil {
			return nil, err
		}
		return types.NewDataPart(dp.Data), nil
	}
	return nil, fmt.Errorf("unknown part kind: %s", kind.Kind)
}
