// MCP Apps bridge: the widget is the MCP *client* of the host. The host keeps
// the iframe hidden (visibility:hidden) and zero-height until we (a) complete
// the ui/initialize handshake and (b) report a size. This module ports the
// proven handshake + tools/call round-trip from the v22 experiment, typed.
//
// Param keys matter: the host's McpUiInitializeRequestSchema expects
// `appCapabilities` / `appInfo` / `protocolVersion` (NOT base-MCP
// `capabilities` / `clientInfo`) and silently drops the wrong keys.

const PROTOCOL_VERSION = '2026-01-26';

// The host's color theme. Surfaced from the ui/initialize result's hostContext
// and kept current via the host-context-changed notification. This is the ONLY
// reliable theme signal in the sandboxed iframe: prefers-color-scheme reports
// the OS/browser setting, NOT the host app's (Claude/ChatGPT) in-app dark mode,
// which users routinely set independently of their OS.
export type TTheme = 'light' | 'dark';

// The subset of the MCP Apps HostContext we consume. The host MAY omit any of
// these. `styles.variables` carries host CSS custom properties; we read a
// background var from it to infer the theme when `theme` itself is absent.
export type THostContext = {
    readonly theme?: TTheme;
    readonly styles?: {
        readonly variables?: Record<string, string | undefined>;
    };
};

type TJsonRpc = {
    readonly jsonrpc: '2.0';
    readonly id?: number;
    readonly method?: string;
    readonly params?: unknown;
    readonly result?: unknown;
    readonly error?: unknown;
};

// Shape of a host-proxied CallToolResult (the subset we read).
type TCallToolResult = {
    readonly structuredContent?: unknown;
    readonly content?: unknown;
    readonly isError?: boolean;
};

const post = (msg: TJsonRpc): void => window.parent.postMessage(msg, '*');

// Report current content height so the host can size the frame. Measures the
// app root, NOT document.documentElement: the root html element fills the
// host-set iframe viewport, so its scrollHeight stays clamped to the previous
// (larger) height and never shrinks when switching to a shorter view. #root has
// auto height, so its bounding box is the true content height and shrinks too.
export const reportSize = (): void => {
    const root = document.getElementById('root');
    const height = root
        ? Math.ceil(root.getBoundingClientRect().height)
        : document.documentElement.scrollHeight;
    post({
        jsonrpc: '2.0',
        method: 'ui/notifications/size-changed',
        params: {height},
    });
};

// --- tools/call correlation ------------------------------------------------
// A widget->server tool call is a JSON-RPC request whose response the host
// proxies back with the same `id`. We track in-flight calls in a module-level
// map so the single message listener (owned by connectBridge) can resolve them.
// Ids start at 100 to never collide with the id:1 ui/initialize request.
type TPending = {
    readonly resolve: (result: TCallToolResult) => void;
    readonly reject: (err: Error) => void;
};
const pending = new Map<number, TPending>();
let nextId = 100;

// Host capabilities advertised in the ui/initialize result. `serverTools` gates
// whether the host will honour app-initiated tools/call at all.
let hostCapabilities: Record<string, unknown> = {};
export const serverToolsSupported = (): boolean =>
    hostCapabilities.serverTools === true;

// `openLinks` gates whether the host will honour ui/open-link (open a URL in the
// user's real browser tab — the only way a download/inline view escapes the
// sandboxed iframe).
export const openLinksSupported = (): boolean =>
    hostCapabilities.openLinks === true;

// --- theme ----------------------------------------------------------------
// The host theme, populated once the ui/initialize handshake resolves (or a
// host-context-changed notification arrives). null until the host speaks.
let hostTheme: TTheme | null = null;
// True once the host has supplied a theme by ANY means. The matchMedia fallback
// (an OS-level signal that can disagree with the host's in-app theme) defers to
// the host once this flips, so a stray OS change never clobbers the real theme.
let hostSpoke = false;
export const hostHasSpoken = (): boolean => hostSpoke;

const prefersDark = (): boolean =>
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-color-scheme: dark)').matches;

// Relative luminance of a CSS color string (hex / rgb[a]); used to classify a
// host-provided background variable as light or dark. Returns null when the
// color can't be parsed, so the caller can fall through to the next signal.
const luminanceOf = (color: string): number | null => {
    const hex = color.trim().match(/^#([0-9a-f]{3}|[0-9a-f]{6})$/i);
    const rgb = color.match(/rgba?\(\s*(\d+)\D+(\d+)\D+(\d+)/i);
    const channels = hex
        ? (hex[1].length === 3
              ? [...hex[1]].map((c) => c + c)
              : hex[1].match(/.{2}/g)!
          ).map((h) => parseInt(h, 16))
        : rgb
          ? [Number(rgb[1]), Number(rgb[2]), Number(rgb[3])]
          : null;
    if (!channels) {
        return null;
    }
    const [r, g, b] = channels;
    return (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255;
};

// Infer a theme from the host's style variables when `hostContext.theme` is
// absent: a dark background variable means dark. Returns null if no usable
// background var is present (caller falls through to matchMedia / default).
const themeFromStyles = (ctx: THostContext): TTheme | null => {
    const vars = ctx.styles?.variables ?? {};
    const bg = Object.entries(vars).find(
        ([k]) =>
            /background|surface|^--?bg/i.test(k) && !/foreground|text/i.test(k),
    )?.[1];
    if (!bg) {
        return null;
    }
    const lum = luminanceOf(bg);
    return lum == null ? null : lum < 0.5 ? 'dark' : 'light';
};

// Best-effort theme for the very first paint, before the handshake lands: the
// host value if it somehow already arrived, else the OS preference (last-resort,
// since it can disagree with the host), else light.
export const detectInitialTheme = (): TTheme =>
    hostTheme ?? (prefersDark() ? 'dark' : 'light');

// Resolve + record the theme the host advertised in a (partial) host context,
// preferring the explicit `theme` field and falling back to style-variable
// luminance. Marks the host as having spoken and fires the handler on any hit.
const applyHostContext = (
    ctx: THostContext,
    onThemeChange?: (theme: TTheme) => void,
): void => {
    const resolved =
        ctx.theme === 'light' || ctx.theme === 'dark'
            ? ctx.theme
            : themeFromStyles(ctx);
    if (!resolved) {
        return;
    }
    hostTheme = resolved;
    hostSpoke = true;
    onThemeChange?.(resolved);
};

const CALL_TIMEOUT_MS = 10_000;

// Call an MCP tool over the bridge and resolve with the host-proxied
// CallToolResult. Only tools the host exposes to apps (`_meta.ui.visibility`
// includes 'app') are callable; others reject. Requires connectBridge() to have
// run first so the response listener is live.
export const callTool = (
    name: string,
    args: Record<string, unknown> = {},
): Promise<TCallToolResult> =>
    new Promise<TCallToolResult>((resolve, reject) => {
        const id = nextId++;
        pending.set(id, {resolve, reject});
        post({
            jsonrpc: '2.0',
            id,
            method: 'tools/call',
            params: {name, arguments: args},
        });
        setTimeout(() => {
            if (pending.has(id)) {
                pending.delete(id);
                reject(new Error(`tools/call "${name}" timed out (10s)`));
            }
        }, CALL_TIMEOUT_MS);
    });

// Ask the host to open a URL in the user's browser (a real tab, outside the
// sandboxed iframe). This is a ui/open-link REQUEST — the host acks with an
// empty result, which we correlate by id through the same pending map as
// callTool. Gated by the openLinks host capability; check openLinksSupported()
// before calling. Resolves when the host acks the open.
export const openLink = (url: string): Promise<void> =>
    new Promise<void>((resolve, reject) => {
        const id = nextId++;
        pending.set(id, {resolve: () => resolve(), reject});
        post({
            jsonrpc: '2.0',
            id,
            method: 'ui/open-link',
            params: {url},
        });
        setTimeout(() => {
            if (pending.has(id)) {
                pending.delete(id);
                reject(new Error('ui/open-link timed out (10s)'));
            }
        }, CALL_TIMEOUT_MS);
    });

// Notifications the host pushes INTO the view after the handshake. tool-input
// carries the invoking tool's arguments ({arguments}); tool-result carries its
// CallToolResult ({structuredContent, ...}). The host sends no tool name, so a
// view shared by several tools discriminates on the result's own payload. (The
// host may also send tool-input-partial while the model streams arguments; we
// ignore it and act only on the complete tool-input.)
export type TBridgeHandlers = {
    readonly onToolInput?: (args: Record<string, unknown>) => void;
    readonly onToolResult?: (structuredContent: unknown) => void;
    // Fired with the host theme on handshake and on every host-context-changed
    // that carries one (or an inferrable style background).
    readonly onThemeChange?: (theme: TTheme) => void;
};

// Drive the handshake to completion: send ui/initialize (with retries, since
// the host may not be listening on the very first tick), and on the host's
// result reply with ui/notifications/initialized so the frame is revealed. The
// same listener resolves any in-flight tools/call by id and forwards the host's
// tool-input / tool-result notifications to the optional handlers.
export const connectBridge = (
    appName: string,
    handlers: TBridgeHandlers = {},
): void => {
    let initialized = false;
    const initRequest: TJsonRpc = {
        jsonrpc: '2.0',
        id: 1,
        method: 'ui/initialize',
        params: {
            appCapabilities: {},
            appInfo: {name: appName, version: '0.0.1'},
            protocolVersion: PROTOCOL_VERSION,
        },
    };

    window.addEventListener('message', (event) => {
        const data = event.data as TJsonRpc | undefined;
        if (!data || data.jsonrpc !== '2.0') {
            return;
        }
        // Response to our ui/initialize -> confirm initialized + read caps and
        // the host context (theme).
        if (data.id === 1 && data.result) {
            initialized = true;
            const result = data.result as {
                readonly hostCapabilities?: Record<string, unknown>;
                readonly hostContext?: THostContext;
            };
            hostCapabilities = result.hostCapabilities ?? {};
            applyHostContext(result.hostContext ?? {}, handlers.onThemeChange);
            post({
                jsonrpc: '2.0',
                method: 'ui/notifications/initialized',
                params: {},
            });
            reportSize();
            return;
        }
        // Host-pushed context change (theme / viewport / locale) -> re-resolve
        // the theme. Params are a PARTIAL host context; we act only on theme.
        if (data.method === 'ui/notifications/host-context-changed') {
            applyHostContext(
                (data.params ?? {}) as THostContext,
                handlers.onThemeChange,
            );
            return;
        }
        // Host-pushed tool input -> forward the arguments to the view.
        if (data.method === 'ui/notifications/tool-input') {
            const params = (data.params ?? {}) as {
                readonly arguments?: Record<string, unknown>;
            };
            handlers.onToolInput?.(params.arguments ?? {});
            return;
        }
        // Host-pushed tool result -> forward its structuredContent to the view.
        if (data.method === 'ui/notifications/tool-result') {
            const params = (data.params ?? {}) as {
                readonly structuredContent?: unknown;
            };
            handlers.onToolResult?.(params.structuredContent);
            return;
        }
        // Response to a tools/call we issued -> resolve/reject by id.
        if (data.id != null && pending.has(data.id)) {
            const p = pending.get(data.id);
            pending.delete(data.id);
            if (!p) {
                return;
            }
            if (data.error) {
                p.reject(new Error(JSON.stringify(data.error)));
                return;
            }
            p.resolve((data.result ?? {}) as TCallToolResult);
        }
    });

    const fire = (attempt: number): void => {
        if (initialized) {
            return;
        }
        post(initRequest);
        if (attempt < 6) {
            setTimeout(() => fire(attempt + 1), attempt * 120);
        }
    };
    fire(1);
    reportSize();
};
