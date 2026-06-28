import '../index.css';
import {render} from 'preact';
import {useEffect, useMemo, useRef, useState} from 'preact/hooks';
import {
    useApp,
    useDocumentTheme,
    useHostStyleVariables,
} from '@modelcontextprotocol/ext-apps/react';
import {
    ArcElement,
    BarController,
    BarElement,
    CategoryScale,
    Chart,
    DoughnutController,
    Filler,
    Legend,
    LinearScale,
    LineController,
    LineElement,
    PieController,
    PointElement,
    Tooltip,
} from 'chart.js';
import {Loader2, TriangleAlert} from 'lucide-react';
import {Card} from '@/components/ui/card';
import {
    buildCartesianConfig,
    buildProportionalConfig,
    type CartesianInput,
    type ProportionalInput,
    type Row,
    type TTheme,
} from './charts';

// Register only the controllers/elements the two chart families use; tree-shaking
// drops the rest. Filler backs area series; the scales/elements cover bar + line;
// Arc covers pie + doughnut.
Chart.register(
    BarController,
    BarElement,
    LineController,
    LineElement,
    PointElement,
    PieController,
    DoughnutController,
    ArcElement,
    CategoryScale,
    LinearScale,
    Filler,
    Tooltip,
    Legend,
);

// The two render tools the host can invoke for this widget. The tool-result's
// structuredContent.tool discriminates which mapping the tool-input carries.
type TToolName = 'render_cartesian_series_chart' | 'render_proportional_chart';

// fetch_chart_data's structured result: the rows the chart plots.
type TChartData = {readonly columns?: string[]; readonly rows?: Row[]};

// readRows pulls the rows out of a host-proxied CallToolResult, failing soft to
// [] so a malformed payload renders an empty chart rather than crashing.
const readRows = (result: {structuredContent?: unknown}): Row[] => {
    const rows = (result.structuredContent as TChartData | undefined)?.rows;
    return Array.isArray(rows) ? rows : [];
};

// queryKeyOf reads the result_key out of the render tool's forwarded arguments.
const queryKeyOf = (args: Record<string, unknown> | null): string | null => {
    const key = args?.query_key;
    return typeof key === 'string' ? key : null;
};

const App = () => {
    const canvasRef = useRef<HTMLCanvasElement | null>(null);
    const chartRef = useRef<Chart | null>(null);

    // The render tool's arguments (the axis/series or value/layer mapping) and
    // the discriminator naming which chart to draw, both pushed by the host.
    const [toolArgs, setToolArgs] = useState<Record<string, unknown> | null>(
        null,
    );
    const [tool, setTool] = useState<TToolName | null>(null);
    const [rows, setRows] = useState<Row[] | null>(null);
    const [error, setError] = useState<string | null>(null);

    // Connect to the host and wire the input/result handlers before the
    // handshake completes (onAppCreated runs pre-connect).
    const {app, isConnected} = useApp({
        appInfo: {name: 'tablekit-chart-renderer', version: '0.1.0'},
        capabilities: {},
        onAppCreated: (created) => {
            created.ontoolinput = (params) =>
                setToolArgs(params.arguments ?? {});
            created.ontoolresult = (params) => {
                const name = (
                    params.structuredContent as {tool?: string} | undefined
                )?.tool;
                if (
                    name === 'render_cartesian_series_chart' ||
                    name === 'render_proportional_chart'
                ) {
                    setTool(name);
                }
            };
        },
    });

    // Mirror host style variables + theme onto the document so shadcn tokens and
    // the `.dark` variant track the host app.
    useHostStyleVariables(app);
    const theme = useDocumentTheme() as TTheme;

    // Once connected and we know which query to chart, load its full result over
    // the bridge via the app-only fetch_chart_data tool.
    const queryKey = queryKeyOf(toolArgs);
    useEffect(() => {
        if (!app || !isConnected || !queryKey) {
            return;
        }
        let cancelled = false;
        setError(null);
        app.callServerTool({
            name: 'fetch_chart_data',
            arguments: {query_key: queryKey},
        })
            .then((result) => {
                if (!cancelled) {
                    setRows(readRows(result));
                }
            })
            .catch((e: unknown) => {
                if (!cancelled) {
                    setError(e instanceof Error ? e.message : String(e));
                }
            });
        return () => {
            cancelled = true;
        };
    }, [app, isConnected, queryKey]);

    // Infer the chart family: the discriminator if the host sent it, otherwise
    // the shape of the arguments (x/y => cartesian, value_prop => proportional).
    const chartKind = useMemo<'cartesian' | 'proportional' | null>(() => {
        if (tool === 'render_cartesian_series_chart') return 'cartesian';
        if (tool === 'render_proportional_chart') return 'proportional';
        if (!toolArgs) return null;
        if ('value_prop' in toolArgs) return 'proportional';
        if ('x' in toolArgs) return 'cartesian';
        return null;
    }, [tool, toolArgs]);

    // (Re)draw whenever data, mapping or theme changes. Chart.js owns the canvas,
    // so recreate the instance rather than mutate it in place.
    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas || !rows || !toolArgs || !chartKind) {
            return;
        }
        chartRef.current?.destroy();
        const config =
            chartKind === 'cartesian'
                ? buildCartesianConfig(
                      toolArgs as unknown as CartesianInput,
                      rows,
                      theme,
                  )
                : buildProportionalConfig(
                      toolArgs as unknown as ProportionalInput,
                      rows,
                      theme,
                  );
        chartRef.current = new Chart(canvas, config);
        return () => chartRef.current?.destroy();
    }, [rows, toolArgs, chartKind, theme]);

    return (
        <Card className="m-2 p-4">
            {error ? (
                <div className="flex items-center gap-2 text-sm text-destructive">
                    <TriangleAlert size={16} />
                    {error}
                </div>
            ) : !rows ? (
                <div className="flex h-64 items-center justify-center gap-2 text-sm text-muted-foreground">
                    <Loader2 size={16} className="animate-spin" />
                    Loading chart data…
                </div>
            ) : (
                <div className="relative h-72">
                    <canvas ref={canvasRef} />
                </div>
            )}
        </Card>
    );
};

const root = document.getElementById('root');
if (root) {
    render(<App />, root);
}
