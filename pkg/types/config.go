package types

import "time"

const (
	DefaultWSUrl1         = "wss://hag.cloud.huawei.com/openclaw/v1/ws/link"
	DefaultWSUrl2         = "wss://116.63.174.231/openclaw/v1/ws/link"
	ProtocolHeartbeat     = 30 * time.Second
	AppHeartbeat          = 20 * time.Second
	HeartbeatTimeout      = 90 * time.Second
	StableThreshold       = 10 * time.Second
	DefaultReconnectDelay = 10 * time.Second
	ReconnectMaxDelay     = 60 * time.Second
	MaxReconnectAttempts  = 50
	ConnectionTimeout     = 30 * time.Second
)

type Config struct {
	AK              string
	SK              string
	AgentID         string
	WSUrl1          string
	WSUrl2          string
	EnableStreaming bool
	ReconnectDelay  time.Duration
	SingleServer    bool // 只连接 server1，避免同一 agentID 多连接
}

func DefaultConfig() *Config {
	return &Config{
		WSUrl1:          DefaultWSUrl1,
		WSUrl2:          DefaultWSUrl2,
		EnableStreaming: true,
		ReconnectDelay:  DefaultReconnectDelay,
	}
}

func (c *Config) Validate() error {
	if c.AK == "" {
		return &XiaoYiError{Code: "CONFIG_INVALID", Message: "AK is required"}
	}
	if c.SK == "" {
		return &XiaoYiError{Code: "CONFIG_INVALID", Message: "SK is required"}
	}
	if c.AgentID == "" {
		return &XiaoYiError{Code: "CONFIG_INVALID", Message: "AgentID is required"}
	}
	return nil
}

func (c *Config) ApplyDefaults() {
	if c.WSUrl1 == "" {
		c.WSUrl1 = DefaultWSUrl1
	}
	if c.WSUrl2 == "" {
		c.WSUrl2 = DefaultWSUrl2
	}
	if c.ReconnectDelay == 0 {
		c.ReconnectDelay = DefaultReconnectDelay
	}
}
