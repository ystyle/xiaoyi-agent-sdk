/**
 * SessionManager - 统一的会话状态管理
 *
 * 职责：
 * 1. 管理会话的生命周期和状态
 * 2. 追踪异步任务（SubAgent）状态
 * 3. 提供清晰的状态查询接口
 * 4. 管理超时和定时器
 */
export interface FinalMessageOptions {
    text?: string;
    isFinal?: boolean;
}
export interface SessionState {
    sessionId: string;
    taskId: string;
    startTime: number;
    accumulatedText: string;
    hasSentFinal: boolean;
    hasAsyncActivity: boolean;
    lastDeliverTime: number;
    gracePeriodTimeout: NodeJS.Timeout | null;
    maxTimeoutId: NodeJS.Timeout | null;
}
export declare class SessionManager {
    private sessions;
    private runtime;
    /**
     * 初始化会话
     */
    initSession(sessionId: string, taskId: string): void;
    /**
     * 更新会话活动时间
     * 每次调用 deliver 时都应调用此方法
     */
    updateActivity(sessionId: string, text: string): void;
    /**
     * 检测是否有异步活动
     * 基于 queuedFinal 和 tool 调用
     */
    detectAsyncActivity(queuedFinal: boolean, counts: any): boolean;
    /**
     * 判断是否为 final 响应
     */
    isFinalResponse(info: any, payload: any): boolean;
    /**
     * 发送 final 消息并标记会话完成
     */
    sendFinal(sessionId: string, conn: any, taskId: string, options: FinalMessageOptions, runtime: any): Promise<void>;
    /**
     * 启动 grace period（宽限期）
     */
    startGracePeriod(sessionId: string, config: {
        onSubAgentOutput: () => void;
        onTimeout: () => void;
    }): void;
    /**
     * 处理 SubAgent 输出（延长宽限期）
     */
    handleSubAgentOutput(sessionId: string): void;
    /**
     * 清除该会话的所有定时器
     */
    clearTimeouts(sessionId: string): void;
    /**
     * 标记会话为完成（不发送 final）
     * 用于简单对话场景，无需 grace period
     */
    markCompleted(sessionId: string): void;
    /**
     * 获取会话状态
     */
    getSessionState(sessionId: string): SessionState | undefined;
    /**
     * 清理会话（移除）
     */
    cleanup(sessionId: string): void;
}
