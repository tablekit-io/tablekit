// Host bridge: the chart widget runs inside two different, incompatible host
// runtimes, and this module hides that difference behind one interface.
//
//   - MCP-UI hosts (Claude, MCP Inspector, …) speak the @modelcontextprotocol
//     /ext-apps postMessage protocol: the widget is an MCP client that hand-
//     shakes over a PostMessageTransport, receives the render tool's arguments
//     via an `ontoolinput` notification, and calls other tools with
//     `app.callServerTool`.
//   - ChatGPT's Apps SDK injects a `window.openai` global instead: the render
//     tool's arguments arrive as `window.openai.toolInput`, the structured
//     result as `window.openai.toolOutput`, tools are called with
//     `window.openai.callTool`, and updates are announced via the
//     `openai:set_globals` DOM event.
//
// The two are unified into `HostBridge` so the chart UI depends only on this
// interface, never on a concrete host SDK. Adding a third host later is a new
// adapter here, not a change to the rendering code.
import {useCallback, useEffect, useState, useSyncExternalStore} from 'react';
import {
    useApp,
    useDocumentTheme,
    useHostStyleVariables,
} from '@modelcontextprotocol/ext-apps/react';

// The result shape both hosts' tool calls share: only structuredContent matters
// to this widget (fetch_chart_data's rows, get_export_url's signed url).
export type ToolResult = {structuredContent?: unknown};

// HostBridge is the single surface the chart UI talks to, regardless of host.
export interface HostBridge {
    // True once the host has connected/hydrated enough to call tools. Until then
    // the widget shows its loading state rather than firing a fetch that would
    // race the handshake.
    ready: boolean;
    // The render tool's arguments: query_key plus the axis/series mapping.
    toolArgs: Record<string, unknown> | null;
    // The discriminator from the render tool's structured result ({tool: name}).
    toolName: string | null;
    theme: 'light' | 'dark';
    // Whether the host can open external links (drives the export button).
    canExport: boolean;
    // Invoke another MCP tool (fetch_chart_data, get_export_url) over the host.
    callTool: (name: string, args: Record<string, unknown>) => Promise<ToolResult>;
    // Open a URL in the user's real browser (the host owns the navigation).
    openLink: (url: string) => void;
}

// --- ChatGPT (window.openai) adapter -------------------------------------

// The subset of ChatGPT's Apps SDK runtime this widget uses. The real global
// carries more; we type only what we touch.
interface OpenAiGlobals {
    toolInput?: Record<string, unknown>;
    toolOutput?: unknown;
    theme?: 'light' | 'dark';
    callTool: (name: string, args: Record<string, unknown>) => Promise<ToolResult>;
    openExternal: (opts: {href: string; redirectUrl?: boolean}) => void;
}

const SET_GLOBALS_EVENT = 'openai:set_globals';

const openAi = (): OpenAiGlobals | undefined =>
    (globalThis as {openai?: OpenAiGlobals}).openai;

// hasOpenAiHost reports whether the ChatGPT Apps SDK runtime is present. The host
// is fixed for a widget's lifetime, so this is read once at mount to pick the
// adapter — never inside a hook, which would risk violating the rules of hooks.
export const hasOpenAiHost = (): boolean =>
    typeof globalThis !== 'undefined' && 'openai' in globalThis;

// useOpenAiGlobal subscribes a component to one window.openai global, re-rendering
// whenever ChatGPT dispatches openai:set_globals (its only change signal).
function useOpenAiGlobal<K extends keyof OpenAiGlobals>(
    key: K,
): OpenAiGlobals[K] | undefined {
    return useSyncExternalStore(
        (onChange) => {
            window.addEventListener(SET_GLOBALS_EVENT, onChange, {passive: true});
            return () => window.removeEventListener(SET_GLOBALS_EVENT, onChange);
        },
        () => openAi()?.[key],
    );
}

function useChatGptHost(): HostBridge {
    const toolInput = useOpenAiGlobal('toolInput');
    const toolOutput = useOpenAiGlobal('toolOutput');
    const theme = useOpenAiGlobal('theme') ?? 'light';

    // Mirror ChatGPT's theme onto the document so the widget's `.dark` variant
    // tracks the host (MCP-UI hosts get this via useHostStyleVariables instead).
    useEffect(() => {
        document.documentElement.classList.toggle('dark', theme === 'dark');
    }, [theme]);

    const callTool = useCallback(
        (name: string, args: Record<string, unknown>): Promise<ToolResult> => {
            const api = openAi();
            return api
                ? api.callTool(name, args)
                : Promise.reject(new Error('ChatGPT host unavailable'));
        },
        [],
    );
    const openLink = useCallback((url: string) => {
        openAi()?.openExternal({href: url});
    }, []);

    return {
        // window.openai is present synchronously in ChatGPT; toolInput may lag by
        // one set_globals tick, which the fetch effect handles via its queryKey
        // guard, so "ready" simply means the runtime exists.
        ready: hasOpenAiHost(),
        toolArgs: toolInput ?? null,
        toolName: (toolOutput as {tool?: string} | undefined)?.tool ?? null,
        theme,
        canExport: typeof openAi()?.openExternal === 'function',
        callTool,
        openLink,
    };
}

// --- MCP-UI (ext-apps) adapter -------------------------------------------

function useMcpUiHost(): HostBridge {
    const [toolArgs, setToolArgs] = useState<Record<string, unknown> | null>(null);
    const [toolName, setToolName] = useState<string | null>(null);

    // Register the notification handlers in onAppCreated (before connect) so no
    // tool-input/result notification is missed during the handshake.
    const {app, isConnected} = useApp({
        appInfo: {name: 'tablekit-chart-renderer', version: '0.1.0'},
        capabilities: {},
        onAppCreated: (created) => {
            created.ontoolinput = (params) => setToolArgs(params.arguments ?? {});
            created.ontoolresult = (params) => {
                const name = (params.structuredContent as {tool?: string} | undefined)
                    ?.tool;
                if (name) {
                    setToolName(name);
                }
            };
        },
    });

    useHostStyleVariables(app);
    const theme = useDocumentTheme();

    const callTool = useCallback(
        (name: string, args: Record<string, unknown>): Promise<ToolResult> =>
            app
                ? app.callServerTool({name, arguments: args})
                : Promise.reject(new Error('MCP host unavailable')),
        [app],
    );
    const openLink = useCallback(
        (url: string) => {
            void app?.openLink({url});
        },
        [app],
    );

    return {
        ready: isConnected,
        toolArgs,
        toolName,
        theme: theme === 'dark' ? 'dark' : 'light',
        // The host opens links in the user's real browser; only offer export when
        // it advertises that capability.
        canExport: isConnected && !!app?.getHostCapabilities()?.openLinks,
        callTool,
        openLink,
    };
}

// useHost selects the adapter for the current host. `hasOpenAiHost()` is a stable
// per-page fact, so this branch is taken once and never flips between renders —
// each render calls exactly one host hook, satisfying the rules of hooks.
export function useHost(): HostBridge {
    if (hasOpenAiHost()) {
        // eslint-disable-next-line react-hooks/rules-of-hooks
        return useChatGptHost();
    }
    // eslint-disable-next-line react-hooks/rules-of-hooks
    return useMcpUiHost();
}
