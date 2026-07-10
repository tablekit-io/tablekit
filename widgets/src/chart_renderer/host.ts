// Host bridge: the chart widget is the MCP *client* of its host. Both Claude and
// ChatGPT are MCP-UI postMessage hosts — the widget handshakes over postMessage,
// receives the render tool's arguments via an `ontoolinput` notification, and
// calls other tools with `callServerTool`.
//
// We use @modelcontextprotocol/ext-apps' `App` for the protocol, but drive it
// with our OWN transport (ResilientTransport) instead of the SDK's built-in
// PostMessageTransport, because the built-in one fails to connect to ChatGPT:
//   1. It ignores any message whose `event.source !== window.parent`. ChatGPT
//      proxies host messages through a sandbox frame whose window ref isn't our
//      parent, so the handshake response is dropped. We accept any source.
//   2. It fires `ui/initialize` once; if the host isn't listening on the first
//      tick (ChatGPT isn't) the request is lost. We re-fire with backoff.
//   3. It strictly zod-validates the host's initialize response; a missing field
//      throws. We backfill defaults so validation passes.
//
// This module is INSTRUMENTED: it logs every step to the console (console.debug/
// info) so a stuck handshake can be diagnosed from the host's devtools. The logs
// are verbose on purpose — this is the dev build.
import {useCallback, useEffect, useRef, useState} from 'react';
import {App} from '@modelcontextprotocol/ext-apps';

const TAG = '[tablekit-widget]';
// Injected by vite (see vite.config.ts): true for the dev/watch build, false for
// production. Declared here so tsc is happy; replaced with a literal at build so
// the logging is dead-code-eliminated from production bundles.
declare const __WIDGET_DEV__: boolean;
// Verbose instrumentation switch. On in dev so the host's console shows the full
// handshake; compiled out of production builds.
const DEBUG = __WIDGET_DEV__;
const log = (...args: unknown[]): void => {
    if (DEBUG) {
        console.debug(TAG, ...args);
    }
};
const info = (...args: unknown[]): void => {
    if (DEBUG) {
        console.info(TAG, ...args);
    }
};

// A postMessage JSON-RPC frame (the subset we read). Loosely typed: the host may
// send requests, responses, or notifications.
type TMessage = {
    readonly jsonrpc?: string;
    readonly id?: unknown;
    readonly method?: string;
    result?: Record<string, unknown>;
    readonly error?: unknown;
    readonly params?: unknown;
};

// The MCP Apps result shape both hosts' tool calls share; only structuredContent
// matters here (fetch_chart_data's rows, get_export_url's signed url).
export type ToolResult = {structuredContent?: unknown};

// HostBridge is the single surface the chart UI talks to.
export interface HostBridge {
    ready: boolean;
    toolArgs: Record<string, unknown> | null;
    toolName: string | null;
    theme: 'light' | 'dark';
    canExport: boolean;
    callTool: (name: string, args: Record<string, unknown>) => Promise<ToolResult>;
    openLink: (url: string) => void;
}

// ResilientTransport is a lenient postMessage transport for the MCP Apps
// handshake (see the module header for why the SDK's own one doesn't reach
// ChatGPT). It implements the MCP SDK Transport surface structurally.
class ResilientTransport {
    onclose?: () => void;
    onerror?: (error: Error) => void;
    onmessage?: (message: TMessage) => void;
    sessionId?: string;
    setProtocolVersion?: (version: string) => void;

    // The id of the in-flight ui/initialize request, so we can recognise (and
    // backfill) its response and stop re-firing once it arrives.
    private initId: unknown = undefined;
    private retryTimer: ReturnType<typeof setInterval> | null = null;
    private settled = false;

    private readonly listener = (event: MessageEvent): void => {
        const data = event.data as TMessage | undefined;
        if (!data || data.jsonrpc !== '2.0') {
            return;
        }
        const sameSource = event.source === window.parent;
        log(
            'recv',
            data.method ?? `response#${String(data.id)}`,
            {sameSource, origin: event.origin},
            data,
        );

        // The init response: stop re-firing and backfill any fields ext-apps'
        // strict schema requires but the host omitted, so connect() doesn't throw.
        if (data.id != null && data.id === this.initId && data.result) {
            info('init response received; backfilling + settling handshake');
            this.settled = true;
            this.stopRetry();
            const result = data.result;
            result.protocolVersion ??= '2026-01-26';
            result.hostInfo ??= {name: 'unknown-host', version: '0.0.0'};
            result.hostCapabilities ??= {};
            result.hostContext ??= {};
        }
        this.onmessage?.(data);
    };

    async start(): Promise<void> {
        info('transport.start — listening for host messages (any source)');
        window.addEventListener('message', this.listener);
    }

    async send(message: TMessage): Promise<void> {
        log('send', message.method ?? `response#${String(message.id)}`, message);
        window.parent.postMessage(message, '*');

        // Re-fire the handshake until the host answers: ChatGPT may not be
        // listening when the first ui/initialize lands.
        if (message.method === 'ui/initialize') {
            this.initId = message.id;
            let attempts = 0;
            this.stopRetry();
            this.retryTimer = setInterval(() => {
                attempts += 1;
                if (this.settled || attempts > 8) {
                    if (!this.settled) {
                        info(`handshake not acked after ${attempts} retries`);
                    }
                    this.stopRetry();
                    return;
                }
                log(`re-firing ui/initialize (attempt ${attempts})`);
                window.parent.postMessage(message, '*');
            }, 180);
        }
    }

    private stopRetry(): void {
        if (this.retryTimer) {
            clearInterval(this.retryTimer);
            this.retryTimer = null;
        }
    }

    async close(): Promise<void> {
        info('transport.close');
        this.stopRetry();
        window.removeEventListener('message', this.listener);
        this.onclose?.();
    }
}

// useHost connects to the host over the resilient bridge and exposes it as a
// HostBridge. The App instance lives in a ref; connection state drives the UI.
export function useHost(): HostBridge {
    // The render arguments (query_key + the axis/series mapping) are read from the
    // tool-RESULT: the render tool echoes them into structuredContent.args, and
    // the host delivers the result against the widget's own call. We deliberately
    // do NOT read the host's tool-input notification — ChatGPT scopes it to the
    // assistant turn, so a same-turn query_database call clobbers it; the result
    // is the only reliably-bound source.
    const [toolArgs, setToolArgs] = useState<Record<string, unknown> | null>(null);
    const [toolName, setToolName] = useState<string | null>(null);
    const [theme, setTheme] = useState<'light' | 'dark'>('light');
    const [ready, setReady] = useState(false);
    const [canExport, setCanExport] = useState(false);
    const appRef = useRef<App | null>(null);

    useEffect(() => {
        info('mounting; creating App + connecting');
        const app = new App(
            {name: 'tablekit-chart-renderer', version: '0.1.0'},
            {},
            {autoResize: true},
        );
        appRef.current = app;

        // Register notification handlers BEFORE connect so none are missed.
        // Logged only for diagnostics — we don't render from tool-input (see the
        // note on the state above).
        app.ontoolinput = (params) => log('ontoolinput (ignored)', params?.arguments);
        app.ontoolresult = (params) => {
            const structured = params?.structuredContent as
                | {tool?: string; args?: Record<string, unknown>}
                | undefined;
            info('ontoolresult', structured);
            if (structured?.tool) {
                setToolName(structured.tool);
            }
            if (structured?.args) {
                setToolArgs(structured.args);
            }
        };
        app.onhostcontextchanged = (context) => {
            info('onhostcontextchanged', {theme: context?.theme});
            if (context?.theme === 'light' || context?.theme === 'dark') {
                setTheme(context.theme);
            }
        };
        app.onerror = (error) => console.error(TAG, 'app.onerror', error);

        const transport = new ResilientTransport();
        // App.connect wants the SDK's Transport type; ResilientTransport is
        // structurally compatible. Cast through a loose signature to avoid
        // pulling the SDK's transport types into the widget.
        const connect = app.connect.bind(app) as (t: unknown) => Promise<void>;

        let cancelled = false;
        connect(transport)
            .then(() => {
                if (cancelled) {
                    return;
                }
                const context = app.getHostContext();
                const capabilities = app.getHostCapabilities();
                info('connected', {
                    hostContext: context,
                    hostCapabilities: capabilities,
                });
                setReady(true);
                if (context?.theme === 'light' || context?.theme === 'dark') {
                    setTheme(context.theme);
                }
                setCanExport(!!capabilities?.openLinks);
            })
            .catch((error: unknown) => {
                if (!cancelled) {
                    console.error(TAG, 'connect failed', error);
                }
            });

        return () => {
            cancelled = true;
            void app.close();
            appRef.current = null;
        };
    }, []);

    // Mirror the host theme onto the document so the widget's `.dark` variant
    // tracks the host (the sandboxed iframe can't read the host's in-app theme).
    useEffect(() => {
        document.documentElement.classList.toggle('dark', theme === 'dark');
    }, [theme]);

    const callTool = useCallback(
        (name: string, args: Record<string, unknown>): Promise<ToolResult> => {
            const app = appRef.current;
            if (!app) {
                return Promise.reject(new Error('host not connected'));
            }
            info('callTool ->', name, args);
            return app
                .callServerTool({name, arguments: args})
                .then((result) => {
                    info('callTool <-', name, result);
                    return result as ToolResult;
                })
                .catch((error: unknown) => {
                    console.error(TAG, 'callTool failed', name, error);
                    throw error;
                });
        },
        [],
    );

    const openLink = useCallback((url: string) => {
        info('openLink', url);
        void appRef.current?.openLink({url});
    }, []);

    return {ready, toolArgs, toolName, theme, canExport, callTool, openLink};
}
