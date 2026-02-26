export interface A2ARequestMessage {
    agentId: string;
    jsonrpc: "2.0";
    id: string;
    method: "message/stream";
    deviceId?: string;
    conversationId?: string;
    sessionId?: string;
    params: {
        id: string;
        sessionId?: string;
        agentLoginSessionId?: string;
        message: {
            kind?: string;
            messageId?: string;
            role: "user" | "agent";
            parts: Array<{
                kind: "text" | "file" | "data";
                text?: string;
                file?: {
                    name: string;
                    mimeType: string;
                    bytes?: string;
                    uri?: string;
                };
                data?: any;
            }>;
        };
    };
}
export interface A2AResponseMessage {
    sessionId: string;
    messageId: string;
    timestamp: number;
    agentId: string;
    sender: {
        id: string;
        name?: string;
        type: "agent";
    };
    content: {
        type: "text" | "image" | "audio" | "video" | "file";
        text?: string;
        mediaUrl?: string;
        fileName?: string;
        fileSize?: number;
        mimeType?: string;
    };
    context?: {
        conversationId?: string;
        threadId?: string;
        replyToMessageId?: string;
    };
    status: "success" | "error" | "processing";
    error?: {
        code: string;
        message: string;
    };
}
export interface A2AJsonRpcResponse {
    jsonrpc: "2.0";
    id: string;
    result?: A2ATaskArtifactUpdateEvent | A2ATaskStatusUpdateEvent | A2AClearContextResult | A2ATasksCancelResult;
    error?: {
        code: number | string;
        message: string;
    };
}
export interface A2ATaskArtifactUpdateEvent {
    taskId: string;
    kind: "artifact-update";
    append?: boolean;
    lastChunk?: boolean;
    final: boolean;
    artifact: {
        artifactId: string;
        parts: Array<{
            kind: "text" | "file" | "data";
            text?: string;
            file?: {
                name: string;
                mimeType: string;
                bytes?: string;
                uri?: string;
            };
            data?: any;
        }>;
    };
}
export interface A2ATaskStatusUpdateEvent {
    taskId: string;
    kind: "status-update";
    final: boolean;
    status: {
        message: {
            role: "agent";
            parts: Array<{
                kind: "text";
                text: string;
            }>;
        };
        state: "submitted" | "working" | "input-required" | "completed" | "canceled" | "failed" | "unknown";
    };
}
export interface A2AClearContextResult {
    status: {
        state: "cleared" | "failed" | "unknown";
    };
}
export interface A2ATasksCancelResult {
    id: string;
    status: {
        state: "canceled" | "failed" | "unknown";
    };
}
export interface A2AWebSocketMessage {
    type: "message" | "heartbeat" | "auth" | "error";
    data: A2ARequestMessage | A2AResponseMessage | any;
}
export type OutboundMessageType = "clawd_bot_init" | "agent_response" | "heartbeat";
export interface OutboundWebSocketMessage {
    msgType: OutboundMessageType;
    agentId: string;
    sessionId?: string;
    taskId?: string;
    msgDetail?: string;
}
export interface A2AClearMessage {
    agentId: string;
    sessionId: string;
    id: string;
    action: "clear";
    timestamp: number;
}
export interface A2ATasksCancelMessage {
    agentId: string;
    sessionId: string;
    id: string;
    action?: "tasks/cancel";
    method?: "tasks/cancel";
    taskId?: string;
    jsonrpc?: "2.0";
    conversationId?: string;
    timestamp?: number;
}
export interface XiaoYiChannelConfig {
    enabled: boolean;
    wsUrl?: string;
    wsUrl1?: string;
    wsUrl2?: string;
    ak: string;
    sk: string;
    agentId: string;
    enableStreaming?: boolean;
}
export interface AuthCredentials {
    ak: string;
    sk: string;
    timestamp: number;
    signature: string;
}
export interface WebSocketConnectionState {
    connected: boolean;
    authenticated: boolean;
    lastHeartbeat: number;
    lastAppHeartbeat: number;
    reconnectAttempts: number;
    maxReconnectAttempts: number;
}
export declare const DEFAULT_WS_URL_1 = "wss://hag.cloud.huawei.com/openclaw/v1/ws/link";
export declare const DEFAULT_WS_URL_2 = "wss://116.63.174.231/openclaw/v1/ws/link";
export interface InternalWebSocketConfig {
    wsUrl1: string;
    wsUrl2: string;
    agentId: string;
    ak: string;
    sk: string;
    enableStreaming?: boolean;
}
export type ServerId = 'server1' | 'server2';
export interface ServerConnectionState {
    connected: boolean;
    ready: boolean;
    lastHeartbeat: number;
    reconnectAttempts: number;
}
