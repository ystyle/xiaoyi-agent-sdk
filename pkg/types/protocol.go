package types

type ServerID string

const (
	Server1 ServerID = "server1"
	Server2 ServerID = "server2"
)

type ServerState struct {
	Connected      bool
	Ready          bool
	LastHeartbeat  int64
	ReconnectCount int
}

type ConnectionState struct {
	Connected      bool
	Authenticated  bool
	LastHeartbeat  int64
	ReconnectCount int
	Server1Ready   bool
	Server2Ready   bool
}

type OutboundMessage struct {
	MsgType   string `json:"msgType"`
	AgentID   string `json:"agentId"`
	SessionID string `json:"sessionId,omitempty"`
	TaskID    string `json:"taskId,omitempty"`
	MsgDetail string `json:"msgDetail,omitempty"`
}

type JsonRpcResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Result  any           `json:"result,omitempty"`
	Error   *JsonRpcError `json:"error,omitempty"`
}

type JsonRpcError struct {
	Code    any    `json:"code"`
	Message string `json:"message"`
}

type ArtifactUpdate struct {
	TaskID    string          `json:"taskId"`
	Kind      string          `json:"kind"`
	Append    bool            `json:"append,omitempty"`
	LastChunk bool            `json:"lastChunk,omitempty"`
	Final     bool            `json:"final"`
	Artifact  ArtifactPayload `json:"artifact"`
}

type ArtifactPayload struct {
	ArtifactID string `json:"artifactId"`
	Parts      []Part `json:"parts"`
}

type StatusUpdate struct {
	TaskID string        `json:"taskId"`
	Kind   string        `json:"kind"`
	Final  bool          `json:"final"`
	Status StatusPayload `json:"status"`
}

type StatusPayload struct {
	Message MessageBody `json:"message"`
	State   string      `json:"state"`
}

type ClearContextResult struct {
	Status struct {
		State string `json:"state"`
	} `json:"status"`
}

type TasksCancelResult struct {
	ID     string `json:"id"`
	Status struct {
		State string `json:"state"`
	} `json:"status"`
}

type PushUpdate struct {
	ID        string         `json:"id"`
	PushID    string         `json:"pushId"`
	PushText  string         `json:"pushText"`
	Kind      string         `json:"kind"`
	Artifacts []PushArtifact `json:"artifacts"`
	Status    PushStatus     `json:"status"`
}

type PushArtifact struct {
	ArtifactID string `json:"artifactId"`
	Parts      []Part `json:"parts"`
}

type PushStatus struct {
	State string `json:"state"`
}
