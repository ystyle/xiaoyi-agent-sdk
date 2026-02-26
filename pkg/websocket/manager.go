package websocket

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ystyle/xiaoyi-agent-sdk/internal/protocol"
	"github.com/ystyle/xiaoyi-agent-sdk/pkg/auth"
	"github.com/ystyle/xiaoyi-agent-sdk/pkg/types"
)

type MessageHandler func(msg *types.A2ARequest)
type ClearHandler func(sessionID string)
type CancelHandler func(sessionID, taskID string)
type ErrorHandler func(serverID types.ServerID, err error)
type StateHandler func(serverID types.ServerID, connected bool)

type Manager struct {
	config *types.Config
	auth   *auth.Auth

	ws1    *websocket.Conn
	ws2    *websocket.Conn
	state1 types.ServerState
	state2 types.ServerState

	ws1Mu sync.Mutex
	ws2Mu sync.Mutex

	connectedTime1 time.Time
	connectedTime2 time.Time

	sessionServerMap map[string]types.ServerID
	mu               sync.RWMutex

	handlers struct {
		message MessageHandler
		clear   ClearHandler
		cancel  CancelHandler
		error   ErrorHandler
		state   StateHandler
	}

	done chan struct{}
	wg   sync.WaitGroup
}

func NewManager(cfg *types.Config) *Manager {
	return &Manager{
		config:           cfg,
		auth:             auth.New(cfg.AK, cfg.SK, cfg.AgentID),
		sessionServerMap: make(map[string]types.ServerID),
		done:             make(chan struct{}),
	}
}

func (m *Manager) OnMessage(h MessageHandler) {
	m.handlers.message = h
}

func (m *Manager) OnClear(h ClearHandler) {
	m.handlers.clear = h
}

func (m *Manager) OnCancel(h CancelHandler) {
	m.handlers.cancel = h
}

func (m *Manager) OnError(h ErrorHandler) {
	m.handlers.error = h
}

func (m *Manager) OnState(h StateHandler) {
	m.handlers.state = h
}

func (m *Manager) Connect(ctx context.Context) error {
	var err1, err2 error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		err1 = m.connectServer1(ctx)
	}()

	go func() {
		defer wg.Done()
		err2 = m.connectServer2(ctx)
	}()

	wg.Wait()

	if err1 != nil && err2 != nil {
		return types.ErrConnectFailed
	}

	go m.startHeartbeat()
	return nil
}

func (m *Manager) connectServer1(ctx context.Context) error {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: isWssWithIP(m.config.WSUrl1),
		},
		HandshakeTimeout: types.ConnectionTimeout,
	}

	headers := m.auth.Headers()
	conn, _, err := dialer.Dial(m.config.WSUrl1, mapToHeader(headers))
	if err != nil {
		if m.handlers.error != nil {
			m.handlers.error(types.Server1, err)
		}
		return err
	}

	m.ws1Mu.Lock()
	m.ws1 = conn
	m.ws1Mu.Unlock()

	m.state1.Connected = true
	m.state1.Ready = true
	m.state1.LastHeartbeat = time.Now().Unix()

	if m.handlers.state != nil {
		m.handlers.state(types.Server1, true)
	}

	initMsg := protocol.BuildInitMessage(m.config.AgentID)
	m.sendToServer1(initMsg)

	go m.readLoop(conn, types.Server1)
	go m.pingLoop(conn, types.Server1, &m.state1, &m.ws1Mu)

	return nil
}

func (m *Manager) connectServer2(ctx context.Context) error {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: isWssWithIP(m.config.WSUrl2),
		},
		HandshakeTimeout: types.ConnectionTimeout,
	}

	headers := m.auth.Headers()
	conn, _, err := dialer.Dial(m.config.WSUrl2, mapToHeader(headers))
	if err != nil {
		if m.handlers.error != nil {
			m.handlers.error(types.Server2, err)
		}
		return err
	}

	m.ws2Mu.Lock()
	m.ws2 = conn
	m.ws2Mu.Unlock()

	m.state2.Connected = true
	m.state2.Ready = true
	m.state2.LastHeartbeat = time.Now().Unix()

	if m.handlers.state != nil {
		m.handlers.state(types.Server2, true)
	}

	initMsg := protocol.BuildInitMessage(m.config.AgentID)
	m.sendToServer2(initMsg)

	go m.readLoop(conn, types.Server2)
	go m.pingLoop(conn, types.Server2, &m.state2, &m.ws2Mu)

	return nil
}

func (m *Manager) sendToServer1(msg *types.OutboundMessage) error {
	m.ws1Mu.Lock()
	defer m.ws1Mu.Unlock()
	if m.ws1 == nil {
		return types.ErrServerNotReady
	}
	data, err := protocol.Marshal(msg)
	if err != nil {
		return err
	}
	fmt.Printf("[SEND server1] %s\n", string(data))
	return m.ws1.WriteMessage(websocket.TextMessage, data)
}

func (m *Manager) sendToServer2(msg *types.OutboundMessage) error {
	m.ws2Mu.Lock()
	defer m.ws2Mu.Unlock()
	if m.ws2 == nil {
		return types.ErrServerNotReady
	}
	data, err := protocol.Marshal(msg)
	if err != nil {
		return err
	}
	fmt.Printf("[SEND server2] %s\n", string(data))
	return m.ws2.WriteMessage(websocket.TextMessage, data)
}

func (m *Manager) readLoop(conn *websocket.Conn, id types.ServerID) {
	m.wg.Add(1)
	defer m.wg.Done()

	for {
		select {
		case <-m.done:
			return
		default:
			_, data, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("[READ ERROR] server=%s error=%v\n", id, err)
				m.handleDisconnect(id)
				return
			}
			m.handleMessage(data, id)
		}
	}
}

func (m *Manager) pingLoop(conn *websocket.Conn, id types.ServerID, state *types.ServerState, wsMu *sync.Mutex) {
	ticker := time.NewTicker(types.ProtocolHeartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			wsMu.Lock()
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				wsMu.Unlock()
				m.handleDisconnect(id)
				return
			}
			wsMu.Unlock()
			if time.Since(time.Unix(state.LastHeartbeat, 0)) > types.HeartbeatTimeout {
				m.handleDisconnect(id)
				return
			}
		}
	}
}

func (m *Manager) startHeartbeat() {
	ticker := time.NewTicker(types.AppHeartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			hb := protocol.BuildHeartbeatMessage(m.config.AgentID)
			m.sendToServer1(hb)
			m.sendToServer2(hb)
		}
	}
}

func (m *Manager) handleMessage(data []byte, sourceServer types.ServerID) {
	msg, err := protocol.ParseA2ARequest(data)
	if err != nil {
		if m.handlers.error != nil {
			m.handlers.error(sourceServer, err)
		}
		return
	}

	sessionID := msg.SessionID()
	if sessionID != "" {
		m.mu.Lock()
		m.sessionServerMap[sessionID] = sourceServer
		m.mu.Unlock()
	}

	if msg.Method == "clearContext" {
		if m.handlers.clear != nil {
			m.handlers.clear(sessionID)
		}
		m.sendClearContextResponse(msg.ID, sessionID, true, sourceServer)
		m.mu.Lock()
		delete(m.sessionServerMap, sessionID)
		m.mu.Unlock()
		return
	}

	if msg.Method == "tasks/cancel" {
		taskID := msg.Params.ID
		if m.handlers.cancel != nil {
			m.handlers.cancel(sessionID, taskID)
		}
		m.sendTasksCancelResponse(msg.ID, sessionID, true, sourceServer)
		return
	}

	if m.handlers.message != nil {
		m.handlers.message(msg)
	}
}

func (m *Manager) handleDisconnect(id types.ServerID) {
	if id == types.Server1 {
		m.ws1Mu.Lock()
		m.state1.Connected = false
		m.state1.Ready = false
		if m.ws1 != nil {
			m.ws1.Close()
			m.ws1 = nil
		}
		m.ws1Mu.Unlock()
	} else {
		m.ws2Mu.Lock()
		m.state2.Connected = false
		m.state2.Ready = false
		if m.ws2 != nil {
			m.ws2.Close()
			m.ws2 = nil
		}
		m.ws2Mu.Unlock()
	}

	if m.handlers.state != nil {
		m.handlers.state(id, false)
	}

	go m.scheduleReconnect(id)
}

func (m *Manager) scheduleReconnect(id types.ServerID) {
	state := &m.state1
	if id == types.Server2 {
		state = &m.state2
	}

	if state.ReconnectCount >= types.MaxReconnectAttempts {
		fmt.Printf("[RECONNECT] %s max attempts reached\n", id)
		return
	}

	delay := m.config.ReconnectDelay * time.Duration(1<<uint(state.ReconnectCount))
	if delay > types.ReconnectMaxDelay {
		delay = types.ReconnectMaxDelay
	}
	state.ReconnectCount++

	fmt.Printf("[RECONNECT] %s attempt %d in %v\n", id, state.ReconnectCount, delay)

	select {
	case <-m.done:
		return
	case <-time.After(delay):
	}

	var err error
	if id == types.Server1 {
		err = m.connectServer1(context.Background())
	} else {
		err = m.connectServer2(context.Background())
	}
	if err != nil {
		fmt.Printf("[RECONNECT] %s failed: %v\n", id, err)
		go m.scheduleReconnect(id)
		return
	}
	fmt.Printf("[RECONNECT] %s success, waiting for stability...\n", id)

	// 等待连接稳定后再重置计数器
	go m.waitForStability(id, state)
}

func (m *Manager) waitForStability(id types.ServerID, state *types.ServerState) {
	timer := time.NewTimer(types.StableThreshold)
	defer timer.Stop()

	select {
	case <-m.done:
		return
	case <-timer.C:
		// 检查连接是否仍然稳定
		if id == types.Server1 {
			m.ws1Mu.Lock()
			stable := m.state1.Connected && m.ws1 != nil
			m.ws1Mu.Unlock()
			if stable {
				fmt.Printf("[STABLE] %s connection stable, resetting reconnect counter\n", id)
				state.ReconnectCount = 0
			}
		} else {
			m.ws2Mu.Lock()
			stable := m.state2.Connected && m.ws2 != nil
			m.ws2Mu.Unlock()
			if stable {
				fmt.Printf("[STABLE] %s connection stable, resetting reconnect counter\n", id)
				state.ReconnectCount = 0
			}
		}
	}
}

func (m *Manager) SendResponse(taskID, sessionID string, response *types.JsonRpcResponse) error {
	m.mu.RLock()
	serverID, ok := m.sessionServerMap[sessionID]
	m.mu.RUnlock()

	if !ok {
		return types.ErrSessionNotFound
	}

	msg := protocol.BuildResponseMessage(m.config.AgentID, sessionID, taskID, response)
	if serverID == types.Server2 {
		return m.sendToServer2(msg)
	}
	return m.sendToServer1(msg)
}

func (m *Manager) sendClearContextResponse(requestID, sessionID string, success bool, target types.ServerID) {
	resp := protocol.BuildClearContextResponse(requestID, success)
	msg := protocol.BuildResponseMessage(m.config.AgentID, sessionID, requestID, resp)

	if target == types.Server2 {
		m.sendToServer2(msg)
	} else {
		m.sendToServer1(msg)
	}
}

func (m *Manager) sendTasksCancelResponse(requestID, sessionID string, success bool, target types.ServerID) {
	resp := protocol.BuildTasksCancelResponse(requestID, success)
	msg := protocol.BuildResponseMessage(m.config.AgentID, sessionID, requestID, resp)

	if target == types.Server2 {
		m.sendToServer2(msg)
	} else {
		m.sendToServer1(msg)
	}
}

func (m *Manager) IsReady() bool {
	m.ws1Mu.Lock()
	ws1ok := m.state1.Ready && m.ws1 != nil
	m.ws1Mu.Unlock()

	m.ws2Mu.Lock()
	ws2ok := m.state2.Ready && m.ws2 != nil
	m.ws2Mu.Unlock()

	return ws1ok || ws2ok
}

func (m *Manager) GetState() *types.ConnectionState {
	return &types.ConnectionState{
		Connected:      m.state1.Connected || m.state2.Connected,
		Authenticated:  m.state1.Connected || m.state2.Connected,
		LastHeartbeat:  maxInt64(m.state1.LastHeartbeat, m.state2.LastHeartbeat),
		ReconnectCount: maxInt(m.state1.ReconnectCount, m.state2.ReconnectCount),
		Server1Ready:   m.state1.Ready,
		Server2Ready:   m.state2.Ready,
	}
}

func (m *Manager) Close() {
	close(m.done)
	m.ws1Mu.Lock()
	if m.ws1 != nil {
		m.ws1.Close()
	}
	m.ws1Mu.Unlock()

	m.ws2Mu.Lock()
	if m.ws2 != nil {
		m.ws2.Close()
	}
	m.ws2Mu.Unlock()

	m.wg.Wait()
}

func isWssWithIP(wsUrl string) bool {
	u, err := url.Parse(wsUrl)
	if err != nil {
		return false
	}
	if u.Scheme != "wss" {
		return false
	}
	return net.ParseIP(u.Hostname()) != nil
}

func mapToHeader(m map[string]string) map[string][]string {
	h := make(map[string][]string)
	for k, v := range m {
		h[k] = []string{v}
	}
	return h
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
