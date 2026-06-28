import '../index.css';
import {useEffect, useMemo, useRef, useState} from 'react';
import {createRoot} from 'react-dom/client';
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
import {Download, Loader2, TriangleAlert} from 'lucide-react';
import {PrismLight as SyntaxHighlighter} from 'react-syntax-highlighter';
import sqlLang from 'react-syntax-highlighter/dist/esm/languages/prism/sql';
import oneDark from 'react-syntax-highlighter/dist/esm/styles/prism/one-dark';
import oneLight from 'react-syntax-highlighter/dist/esm/styles/prism/one-light';
import {Card} from '@/components/ui/card';
import {Button} from '@/components/ui/button';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
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

SyntaxHighlighter.registerLanguage('sql', sqlLang);

// The two render tools the host can invoke for this widget. The tool-result's
// structuredContent.tool discriminates which mapping the tool-input carries.
type TToolName = 'render_cartesian_series_chart' | 'render_proportional_chart';

// fetch_chart_data's structured result: the rows the chart plots plus the SQL.
type TChartData = {
    readonly columns: string[];
    readonly rows: Row[];
    readonly sql: string;
};

// Most rows we put in the DOM for the Table tab. A query can return up to 100k
// rows; rendering them all would lock the iframe, so we cap and note the rest.
const TABLE_MAX_ROWS = 500;

// readChartData pulls the structured result of fetch_chart_data, failing soft to
// empty values so a malformed payload renders empty views rather than crashing.
const readChartData = (result: {structuredContent?: unknown}): TChartData => {
    const data = result.structuredContent as Partial<TChartData> | undefined;
    return {
        columns: Array.isArray(data?.columns) ? data!.columns : [],
        rows: Array.isArray(data?.rows) ? data!.rows : [],
        sql: typeof data?.sql === 'string' ? data!.sql : '',
    };
};

// queryKeyOf reads the result_key out of the render tool's forwarded arguments.
const queryKeyOf = (args: Record<string, unknown> | null): string | null => {
    const key = args?.query_key;
    return typeof key === 'string' ? key : null;
};

// cellText renders a normalized cell value for the Table tab.
const cellText = (value: unknown): string =>
    value == null ? '' : String(value);

const App = () => {
    const canvasRef = useRef<HTMLCanvasElement | null>(null);
    const chartRef = useRef<Chart | null>(null);

    // The render tool's arguments (the axis/series or value/layer mapping) and
    // the discriminator naming which chart to draw, both pushed by the host.
    const [toolArgs, setToolArgs] = useState<Record<string, unknown> | null>(
        null,
    );
    const [tool, setTool] = useState<TToolName | null>(null);
    const [data, setData] = useState<TChartData | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [activeTab, setActiveTab] = useState('chart');
    const [exporting, setExporting] = useState<'csv' | 'json' | null>(null);

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
                    setData(readChartData(result));
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

    // (Re)draw whenever data, mapping, theme or the active tab changes. The
    // canvas only exists while the Chart tab is shown, so recreate on entry and
    // destroy on leave (cheap, and avoids hidden-canvas resize quirks).
    useEffect(() => {
        if (activeTab !== 'chart') {
            return;
        }
        const canvas = canvasRef.current;
        if (!canvas || !data || !toolArgs || !chartKind) {
            return;
        }
        chartRef.current?.destroy();
        const config =
            chartKind === 'cartesian'
                ? buildCartesianConfig(
                      toolArgs as unknown as CartesianInput,
                      data.rows,
                      theme,
                  )
                : buildProportionalConfig(
                      toolArgs as unknown as ProportionalInput,
                      data.rows,
                      theme,
                  );
        chartRef.current = new Chart(canvas, config);
        return () => chartRef.current?.destroy();
    }, [data, toolArgs, chartKind, theme, activeTab]);

    // The host opens links in the user's real browser; only offer export when it
    // advertises that capability.
    const canExport = isConnected && !!app?.getHostCapabilities()?.openLinks;

    const exportAs = async (format: 'csv' | 'json') => {
        if (!app || !queryKey) {
            return;
        }
        setExporting(format);
        try {
            const result = await app.callServerTool({
                name: 'get_export_url',
                arguments: {query_key: queryKey, format},
            });
            const url = (result.structuredContent as {url?: string} | undefined)
                ?.url;
            if (url) {
                await app.openLink({url});
            }
        } catch (e) {
            setError(e instanceof Error ? e.message : String(e));
        } finally {
            setExporting(null);
        }
    };

    if (error) {
        return (
            <Card className="m-2 flex items-center gap-2 p-4 text-sm text-destructive">
                <TriangleAlert size={16} />
                {error}
            </Card>
        );
    }

    if (!data) {
        return (
            <Card className="m-2 flex h-64 items-center justify-center gap-2 p-4 text-sm text-muted-foreground">
                <Loader2 size={16} className="animate-spin" />
                Loading chart data…
            </Card>
        );
    }

    const hiddenRows = Math.max(0, data.rows.length - TABLE_MAX_ROWS);

    return (
        <Card className="m-2 p-4">
            <Tabs value={activeTab} onValueChange={setActiveTab}>
                <div className="mb-3 flex items-center justify-between gap-4">
                    <TabsList>
                        <TabsTrigger value="chart">Chart</TabsTrigger>
                        <TabsTrigger value="table">Table</TabsTrigger>
                        <TabsTrigger value="sql">SQL</TabsTrigger>
                    </TabsList>
                    {canExport && (
                        <div className="flex gap-2">
                            <Button
                                variant="outline"
                                size="sm"
                                disabled={exporting !== null}
                                onClick={() => void exportAs('csv')}>
                                <Download size={14} />
                                CSV
                            </Button>
                            <Button
                                variant="outline"
                                size="sm"
                                disabled={exporting !== null}
                                onClick={() => void exportAs('json')}>
                                <Download size={14} />
                                JSON
                            </Button>
                        </div>
                    )}
                </div>

                <TabsContent value="chart">
                    <div className="relative h-72">
                        <canvas ref={canvasRef} />
                    </div>
                </TabsContent>

                <TabsContent value="table">
                    <div className="max-h-72 overflow-auto rounded-md border border-border">
                        <table className="w-full text-left text-sm">
                            <thead className="sticky top-0 bg-muted">
                                <tr>
                                    {data.columns.map((column) => (
                                        <th
                                            key={column}
                                            className="px-3 py-1.5 font-medium">
                                            {column}
                                        </th>
                                    ))}
                                </tr>
                            </thead>
                            <tbody>
                                {data.rows
                                    .slice(0, TABLE_MAX_ROWS)
                                    .map((row, i) => (
                                        <tr
                                            key={i}
                                            className="border-t border-border">
                                            {data.columns.map((column) => (
                                                <td
                                                    key={column}
                                                    className="px-3 py-1.5">
                                                    {cellText(row[column])}
                                                </td>
                                            ))}
                                        </tr>
                                    ))}
                            </tbody>
                        </table>
                    </div>
                    {hiddenRows > 0 && (
                        <p className="mt-2 text-xs text-muted-foreground">
                            Showing first {TABLE_MAX_ROWS} of {data.rows.length}{' '}
                            rows. Use the export buttons for the full result.
                        </p>
                    )}
                </TabsContent>

                <TabsContent value="sql">
                    <div className="max-h-72 overflow-auto rounded-md border border-border text-sm">
                        <SyntaxHighlighter
                            language="sql"
                            style={theme === 'dark' ? oneDark : oneLight}
                            customStyle={{margin: 0, background: 'transparent'}}
                            wrapLongLines>
                            {data.sql || '-- no SQL available'}
                        </SyntaxHighlighter>
                    </div>
                </TabsContent>
            </Tabs>
        </Card>
    );
};

const root = document.getElementById('root');
if (root) {
    createRoot(root).render(<App />);
}
