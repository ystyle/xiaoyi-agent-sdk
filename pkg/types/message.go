package types

type Message interface {
	TaskID() string
	SessionID() string
	Text() string
	Parts() []Part
}

type Part interface {
	Kind() string
}

type TextPart struct {
	KindField string `json:"kind"`
	Text      string `json:"text"`
}

func NewTextPart(text string) *TextPart {
	return &TextPart{KindField: "text", Text: text}
}

func (p *TextPart) Kind() string { return p.KindField }

type FilePart struct {
	KindField string `json:"kind"`
	File      File   `json:"file"`
}

type File struct {
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
	Bytes    []byte `json:"bytes,omitempty"`
	URI      string `json:"uri,omitempty"`
}

func NewFilePart(name, mimeType, uri string, bytes []byte) *FilePart {
	return &FilePart{
		KindField: "file",
		File: File{
			Name:     name,
			MimeType: mimeType,
			URI:      uri,
			Bytes:    bytes,
		},
	}
}

func (p *FilePart) Kind() string     { return p.KindField }
func (p *FilePart) Name() string     { return p.File.Name }
func (p *FilePart) MimeType() string { return p.File.MimeType }
func (p *FilePart) Bytes() []byte    { return p.File.Bytes }
func (p *FilePart) URI() string      { return p.File.URI }

type DataPart struct {
	KindField string `json:"kind"`
	Data      any    `json:"data"`
}

func NewDataPart(data any) *DataPart {
	return &DataPart{KindField: "data", Data: data}
}

func (p *DataPart) Kind() string { return p.KindField }

type A2ARequest struct {
	JSONRPC        string        `json:"jsonrpc"`
	ID             string        `json:"id"`
	Method         string        `json:"method"`
	AgentID        string        `json:"agentId"`
	DeviceID       string        `json:"deviceId,omitempty"`
	ConversationID string        `json:"conversationId,omitempty"`
	SessionIDField string        `json:"sessionId,omitempty"`
	Params         RequestParams `json:"params"`
}

type RequestParams struct {
	ID                  string      `json:"id"`
	SessionIDField      string      `json:"sessionId,omitempty"`
	AgentLoginSessionID string      `json:"agentLoginSessionId,omitempty"`
	Message             MessageBody `json:"message"`
}

type MessageBody struct {
	Kind      string `json:"kind,omitempty"`
	MessageID string `json:"messageId,omitempty"`
	Role      string `json:"role"`
	Parts     []Part `json:"parts"`
}

func (r *A2ARequest) TaskID() string {
	return r.Params.ID
}

func (r *A2ARequest) SessionID() string {
	if r.Params.SessionIDField != "" {
		return r.Params.SessionIDField
	}
	return r.SessionIDField
}

func (r *A2ARequest) Text() string {
	for _, p := range r.Params.Message.Parts {
		if t, ok := p.(*TextPart); ok {
			return t.Text
		}
	}
	return ""
}

func (r *A2ARequest) Parts() []Part {
	return r.Params.Message.Parts
}

type A2AResponse struct {
	SessionID string           `json:"sessionId"`
	MessageID string           `json:"messageId"`
	Timestamp int64            `json:"timestamp"`
	AgentID   string           `json:"agentId"`
	Sender    Sender           `json:"sender"`
	Content   ResponseContent  `json:"content"`
	Context   *ResponseContext `json:"context,omitempty"`
	Status    string           `json:"status"`
	Error     *ResponseError   `json:"error,omitempty"`
}

type Sender struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Type string `json:"type"`
}

type ResponseContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	MediaURL string `json:"mediaUrl,omitempty"`
	FileName string `json:"fileName,omitempty"`
	FileSize int64  `json:"fileSize,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

type ResponseContext struct {
	ConversationID   string `json:"conversationId,omitempty"`
	ThreadID         string `json:"threadId,omitempty"`
	ReplyToMessageID string `json:"replyToMessageId,omitempty"`
}

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
