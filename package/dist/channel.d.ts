import type { ChannelOutboundContext, OutboundDeliveryResult, ChannelGatewayContext, ChannelMessagingNormalizeTargetContext, ChannelStatusGetAccountStatusContext, OpenClawConfig } from "openclaw";
import { XiaoYiChannelConfig } from "./types";
/**
 * Resolved XiaoYi account configuration (single account mode)
 */
export interface ResolvedXiaoYiAccount {
    accountId: string;
    config: XiaoYiChannelConfig;
}
/**
 * XiaoYi Channel Plugin
 * Implements OpenClaw ChannelPlugin interface for XiaoYi A2A protocol
 * Single account mode only
 */
export declare const xiaoyiPlugin: {
    id: string;
    meta: {
        id: string;
        label: string;
        selectionLabel: string;
        docsPath: string;
        blurb: string;
        aliases: string[];
    };
    capabilities: {
        chatTypes: string[];
        polls: boolean;
        reactions: boolean;
        threads: boolean;
        media: boolean;
        nativeCommands: boolean;
    };
    onboarding: any;
    /**
     * Config adapter - single account mode
     */
    config: {
        listAccountIds: (cfg: OpenClawConfig) => string[];
        resolveAccount: (cfg: OpenClawConfig, accountId?: string | null) => {
            accountId: string;
            config: XiaoYiChannelConfig;
            enabled: boolean;
        };
        defaultAccountId: (cfg: OpenClawConfig) => string;
        isConfigured: (account: any, cfg: OpenClawConfig) => boolean;
        isEnabled: (account: any, cfg: OpenClawConfig) => boolean;
        disabledReason: (account: any, cfg: OpenClawConfig) => string;
        unconfiguredReason: (account: any, cfg: OpenClawConfig) => string;
        describeAccount: (account: any, cfg: OpenClawConfig) => {
            accountId: any;
            name: string;
            enabled: any;
            configured: boolean;
        };
    };
    /**
     * Outbound adapter - send messages
     */
    outbound: {
        deliveryMode: string;
        textChunkLimit: number;
        sendText: (ctx: ChannelOutboundContext) => Promise<OutboundDeliveryResult>;
        sendMedia: (ctx: ChannelOutboundContext) => Promise<OutboundDeliveryResult>;
    };
    /**
     * Gateway adapter - manage connections
     */
    gateway: {
        startAccount: (ctx: ChannelGatewayContext<ResolvedXiaoYiAccount>) => Promise<void>;
        stopAccount: (ctx: ChannelGatewayContext<ResolvedXiaoYiAccount>) => Promise<void>;
    };
    /**
     * Messaging adapter - normalize targets
     */
    messaging: {
        normalizeTarget: (ctx: ChannelMessagingNormalizeTargetContext) => Promise<string>;
    };
    /**
     * Status adapter - health checks
     */
    status: {
        getAccountStatus: (ctx: ChannelStatusGetAccountStatusContext) => Promise<{
            status: string;
            message: string;
        }>;
    };
};
