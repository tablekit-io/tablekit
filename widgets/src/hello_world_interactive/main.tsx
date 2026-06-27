import {render} from 'preact';
import {useEffect, useRef, useState} from 'preact/hooks';
import {defineConfig, install} from '@twind/core';
import presetAutoprefix from '@twind/preset-autoprefix';
import presetTailwind from '@twind/preset-tailwind';
import {
    Chart,
    DoughnutController,
    ArcElement,
    Tooltip,
    Legend,
    type ChartConfiguration,
} from 'chart.js';
import {Loader2, RefreshCw, TriangleAlert} from 'lucide-react';
import {
    callTool,
    connectBridge,
    detectInitialTheme,
    reportSize,
    type TTheme,
} from '../lib/mcp-bridge';

// Runtime Twind: styles are generated in the browser from the class names below
// (no separate stylesheet — exactly what a self-contained single-file widget
// needs). install() also starts the observer that turns class="..." into CSS.
// darkMode:'class' makes `dark:` variants apply when <html> carries the `dark`
// class — the App toggles that class from the host-reported theme (the iframe
// can't read the host app's in-app theme via prefers-color-scheme).
install(
    defineConfig({
        darkMode: 'class',
        presets: [presetAutoprefix(), presetTailwind()],
    }),
);

// Register only the donut subset of Chart.js — the controller, the arc element
// it draws, and the tooltip/legend chrome. Tree-shaking drops everything else.
Chart.register(DoughnutController, ArcElement, Tooltip, Legend);

// One donut slice. Mirrors the hello_world_interactive_data tool's output row.
type TSlice = {readonly label: string; readonly value: number};

// The app-only loader returns {data: TSlice[]} as structuredContent. This is
// the only field we read off the host-proxied CallToolResult.
type TData = {readonly data?: readonly TSlice[]};

// How many slices to ask the loader for. The loader clamps/defaults server-side;
// we just send a friendly count.
const SLICES = 5;

// ---------------------------------------------------------------------------
// TODO(you): pick the color recipe.
//
// This maps each slice label to a fill color for the donut. There's a real
// design choice here: a fixed palette (cycled by index) is simplest and stable;
// a deterministic per-label hue (hash the label -> HSL) keeps a given category
// the same color across re-randomizations; and you may want light/dark to use
// different saturation/lightness so arcs read well on both host themes.
//
// Implement it to return one CSS color per label, same length/order as `labels`.
// ---------------------------------------------------------------------------
function sliceColors(labels: readonly string[], theme: TTheme): string[] {
    return labels.map((_, i) => {
        const hue = Math.round((360 / Math.max(labels.length, 1)) * i);
        const light = theme === 'dark' ? 55 : 60;
        const sat = theme === 'dark' ? 70 : 65;
        return `hsl(${hue} ${sat}% ${light}%)`;
    });
}

// Pull the slice rows out of a callTool result. Fails soft to [] so a malformed
// payload renders an empty (rather than crashed) chart.
const readSlices = (result: {structuredContent?: unknown}): TSlice[] => {
    const data = (result.structuredContent as TData | undefined)?.data;
    return Array.isArray(data) ? data.filter((d) => d && d.label != null) : [];
};

const App = () => {
    const canvasRef = useRef<HTMLCanvasElement | null>(null);
    const chartRef = useRef<Chart<'doughnut'> | null>(null);
    const [theme, setTheme] = useState<TTheme>(detectInitialTheme());
    const [slices, setSlices] = useState<TSlice[] | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Ask the app-only loader for a fresh random dataset over the bridge.
    const load = async () => {
        setLoading(true);
        setError(null);
        try {
            const result = await callTool('hello_world_interactive_data', {
                slices: SLICES,
            });
            setSlices(readSlices(result));
        } catch (e) {
            setError(e instanceof Error ? e.message : String(e));
        } finally {
            setLoading(false);
        }
    };

    // Handshake with the host, then load the first dataset. onThemeChange keeps
    // the <html> dark class + chart colors in sync with the host app's theme.
    useEffect(() => {
        connectBridge('tablekit-hello-world-interactive', {
            onThemeChange: setTheme,
        });
        void load();
    }, []);

    // Reflect the host theme onto <html> so Twind `dark:` variants apply.
    useEffect(() => {
        document.documentElement.classList.toggle('dark', theme === 'dark');
    }, [theme]);

    // (Re)draw the donut whenever data or theme changes. Chart.js owns the
    // canvas, so we recreate the instance rather than mutate it in place.
    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas || !slices) {
            return;
        }
        chartRef.current?.destroy();
        const labels = slices.map((s) => s.label);
        const config: ChartConfiguration<'doughnut'> = {
            type: 'doughnut',
            data: {
                labels,
                datasets: [
                    {
                        data: slices.map((s) => s.value),
                        backgroundColor: sliceColors(labels, theme),
                        borderWidth: 0,
                    },
                ],
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                cutout: '62%',
                plugins: {
                    legend: {
                        position: 'right',
                        labels: {
                            color: theme === 'dark' ? '#e5e7eb' : '#374151',
                        },
                    },
                },
            },
        };
        chartRef.current = new Chart(canvas, config);
        reportSize();
        return () => chartRef.current?.destroy();
    }, [slices, theme]);

    return (
        <div class="p-4 font-sans text-gray-900 dark:text-gray-100">
            <div class="mb-3 flex items-center justify-between gap-4">
                <h1 class="text-base font-semibold">
                    Hello, interactive donut 🍩
                </h1>
                <button
                    type="button"
                    onClick={() => void load()}
                    disabled={loading}
                    class="inline-flex items-center gap-1.5 rounded-md bg-gray-900 px-3 py-1.5 text-sm font-medium text-white hover:bg-gray-700 disabled:opacity-50 dark:bg-gray-100 dark:text-gray-900 dark:hover:bg-gray-300">
                    {loading ? (
                        <Loader2 size={14} class="animate-spin" />
                    ) : (
                        <RefreshCw size={14} />
                    )}
                    Randomize
                </button>
            </div>

            {error ? (
                <div class="flex items-center gap-2 rounded-md bg-red-50 p-3 text-sm text-red-700 dark:bg-red-950 dark:text-red-300">
                    <TriangleAlert size={16} />
                    {error}
                </div>
            ) : (
                <div class="relative h-64">
                    <canvas ref={canvasRef} />
                </div>
            )}
        </div>
    );
};

const root = document.getElementById('root');
if (root) {
    render(<App />, root);
}
