"use strict";
/**
 * SessionManager - 统一的会话状态管理
 *
 * 职责：
 * 1. 管理会话的生命周期和状态
 * 2. 追踪异步任务（SubAgent）状态
 * 3. 提供清晰的状态查询接口
 * 4. 管理超时和定时器
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.SessionManager = void 0;
const runtime_1 = require("./runtime");
// Constants for timeout configuration
const SUBAGENT_GRACE_PERIOD_MS = 30000; // 30秒初始宽限期
const SUBAGENT_MAX_WAIT_MS = 120000; // 最大2分钟
const SUBACTIVITY_TIMEOUT_MS = 5000; // 5秒无活动=空闲
class SessionManager {
    constructor() {
        this.sessions = new Map();
        this.runtime = (0, runtime_1.getXiaoYiRuntime)();
    }
    /**
     * 初始化会话
     */
    initSession(sessionId, taskId) {
        const now = Date.now();
        const session = {
            sessionId,
            taskId,
            startTime: now,
            accumulatedText: "",
            hasSentFinal: false,
            hasAsyncActivity: false,
            lastDeliverTime: now,
            gracePeriodTimeout: null,
            maxTimeoutId: null,
        };
        this.sessions.set(sessionId, session);
        console.log(`\n[SESSION] Initialized session ${sessionId}`);
        console.log(`  Task ID: ${taskId}`);
        console.log(`  Start time: ${now}`);
    }
    /**
     * 更新会话活动时间
     * 每次调用 deliver 时都应调用此方法
     */
    updateActivity(sessionId, text) {
        const session = this.sessions.get(sessionId);
        if (!session) {
            console.warn(`[SESSION] Session ${sessionId} not found for activity update`);
            return;
        }
        const previousLength = session.accumulatedText.length;
        session.accumulatedText = text;
        session.lastDeliverTime = Date.now();
        console.log(`\n[ACTIVITY] Session ${sessionId}`);
        console.log(`  Text length: ${text.length} (added ${text.length - previousLength})`);
        console.log(`  Elapsed: ${Date.now() - session.startTime}ms`);
    }
    /**
     * 检测是否有异步活动
     * 基于 queuedFinal 和 tool 调用
     */
    detectAsyncActivity(queuedFinal, counts) {
        return queuedFinal === true || (counts?.tool || 0) > 0;
    }
    /**
     * 判断是否为 final 响应
     */
    isFinalResponse(info, payload) {
        const kindFinal = info?.kind === "final";
        const statusFinal = payload?.status === "final";
        const payloadQueuedFinal = payload?.queuedFinal === true;
        return kindFinal || statusFinal || payloadQueuedFinal;
    }
    /**
     * 发送 final 消息并标记会话完成
     */
    async sendFinal(sessionId, conn, taskId, options, runtime) {
        const session = this.sessions.get(sessionId);
        if (!session) {
            console.error(`[SESSION] Cannot send final for session ${sessionId}: session not found`);
            return;
        }
        // 检查是否已发送
        if (session.hasSentFinal) {
            console.log(`[SESSION] Final already sent for session ${sessionId}, skipping`);
            return;
        }
        // 标记为已发送
        session.hasSentFinal = true;
        // 清除所有定时器
        this.clearTimeouts(sessionId);
        const response = {
            sessionId,
            messageId: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
            timestamp: Date.now(),
            agentId: this.runtime.getConfig()?.agentId || "",
            sender: {
                id: this.runtime.getConfig()?.agentId || "",
                name: "OpenClaw Agent",
                type: "agent",
            },
            content: {
                type: "text",
                text: options.text || session.accumulatedText,
            },
            status: "success",
        };
        try {
            // 发送 isFinal=true
            await conn.sendResponse(response, taskId, sessionId, options.isFinal !== false, false);
            const elapsed = Date.now() - session.startTime;
            console.log(`\n[SESSION] Final message sent for session ${sessionId}`);
            console.log(`  Total elapsed: ${elapsed}ms`);
            console.log(`  Text length: ${session.accumulatedText.length} chars`);
            console.log(`  isFinal: ${options.isFinal !== false ? options.isFinal : true}`);
        }
        catch (error) {
            console.error(`[SESSION] Failed to send final for session ${sessionId}:`, error);
        }
    }
    /**
     * 启动 grace period（宽限期）
     */
    startGracePeriod(sessionId, config) {
        const session = this.sessions.get(sessionId);
        if (!session) {
            console.warn(`[SESSION] Cannot start grace period for session ${sessionId}: session not found`);
            return;
        }
        if (session.hasSentFinal) {
            console.log(`[SESSION] Final already sent for session ${sessionId}, skipping grace period`);
            return;
        }
        // 清除旧的定时器
        this.clearTimeouts(sessionId);
        console.log(`\n[GRACE-PERIOD] Starting grace period for session ${sessionId}`);
        console.log(`  Grace period: ${SUBAGENT_GRACE_PERIOD_MS}ms`);
        console.log(`  Max wait: ${SUBAGENT_MAX_WAIT_MS}ms`);
        console.log(`  Inactivity timeout: ${SUBACTIVITY_TIMEOUT_MS}ms`);
        // 初始宽限期超时
        session.gracePeriodTimeout = setTimeout(async () => {
            const timeSinceLastDeliver = Date.now() - session.lastDeliverTime;
            if (timeSinceLastDeliver < SUBACTIVITY_TIMEOUT_MS) {
                // 有活动，延长宽限期
                console.log(`\n[GRACE-PERIOD] Activity detected, extending grace period for session ${sessionId}`);
                console.log(`  Time since last deliver: ${timeSinceLastDeliver}ms`);
                session.gracePeriodTimeout = setTimeout(async () => {
                    await config.onTimeout();
                }, SUBACTIVITY_TIMEOUT_MS);
            }
            else {
                // 无活动，发送 final
                console.log(`\n[GRACE-PERIOD] No activity for ${SUBACTIVITY_TIMEOUT_MS}ms, sending final for session ${sessionId}`);
                console.log(`  Time since last deliver: ${timeSinceLastDeliver}ms}`);
                await this.sendFinal(sessionId, this.runtime.getConnection(), session.taskId, {
                    text: session.accumulatedText,
                    isFinal: true,
                }, this.runtime);
                // 标记为已发送
                session.hasSentFinal = true;
                session.gracePeriodTimeout = null;
            }
        }, SUBAGENT_GRACE_PERIOD_MS);
        // 最大超时保护
        session.maxTimeoutId = setTimeout(async () => {
            if (session.gracePeriodTimeout !== null && !session.hasSentFinal) {
                console.log(`\n[GRACE-PERIOD] Max timeout (${SUBAGENT_MAX_WAIT_MS}ms) reached for session ${sessionId}`);
                this.clearTimeouts(sessionId);
                await this.sendFinal(sessionId, this.runtime.getConnection(), session.taskId, {
                    text: session.accumulatedText,
                    isFinal: true,
                }, this.runtime);
                session.hasSentFinal = true;
                session.gracePeriodTimeout = null;
            }
        }, SUBAGENT_MAX_WAIT_MS);
    }
    /**
     * 处理 SubAgent 输出（延长宽限期）
     */
    handleSubAgentOutput(sessionId) {
        const session = this.sessions.get(sessionId);
        if (!session || session.hasSentFinal) {
            return;
        }
        console.log(`\n[SUBAGENT] Output detected for session ${sessionId}`);
        console.log(`  Time since last deliver: ${Date.now() - session.lastDeliverTime}ms`);
        // 清除现有宽限期，重新启动
        if (session.gracePeriodTimeout) {
            clearTimeout(session.gracePeriodTimeout);
            console.log(`[SUBAGENT] Cleared existing grace period timeout`);
        }
        // 启动新的宽限期
        session.gracePeriodTimeout = setTimeout(async () => {
            console.log(`\n[GRACE-PERIOD] Grace period expired, sending final for session ${sessionId}`);
            await this.sendFinal(sessionId, this.runtime.getConnection(), session.taskId, {
                text: session.accumulatedText,
                isFinal: true,
            }, this.runtime);
        }, SUBACTIVITY_TIMEOUT_MS);
    }
    /**
     * 清除该会话的所有定时器
     */
    clearTimeouts(sessionId) {
        const session = this.sessions.get(sessionId);
        if (!session) {
            return;
        }
        if (session.gracePeriodTimeout) {
            clearTimeout(session.gracePeriodTimeout);
            session.gracePeriodTimeout = null;
            console.log(`[CLEANUP] Cleared grace period timeout for session ${sessionId}`);
        }
        if (session.maxTimeoutId) {
            clearTimeout(session.maxTimeoutId);
            session.maxTimeoutId = null;
            console.log(`[CLEANUP] Cleared max timeout for session ${sessionId}`);
        }
    }
    /**
     * 标记会话为完成（不发送 final）
     * 用于简单对话场景，无需 grace period
     */
    markCompleted(sessionId) {
        const session = this.sessions.get(sessionId);
        if (!session) {
            console.warn(`[SESSION] Session ${sessionId} not found for completion`);
            return;
        }
        console.log(`\n[SESSION] Marked session ${sessionId} as completed (no grace period needed)`);
        console.log(`  Total elapsed: ${Date.now() - session.startTime}ms`);
        console.log(`  Final text: ${session.accumulatedText.length} chars`);
    }
    /**
     * 获取会话状态
     */
    getSessionState(sessionId) {
        return this.sessions.get(sessionId);
    }
    /**
     * 清理会话（移除）
     */
    cleanup(sessionId) {
        const session = this.sessions.get(sessionId);
        if (session) {
            this.clearTimeouts(sessionId);
            this.sessions.delete(sessionId);
            console.log(`\n[SESSION] Cleaned up session ${sessionId}`);
            console.log(`  Total elapsed: ${Date.now() - session.startTime}ms`);
            console.log(`  Final text: ${session.accumulatedText.length} chars`);
        }
    }
}
exports.SessionManager = SessionManager;
