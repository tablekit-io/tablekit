import '../index.css';
import {useEffect, useMemo, useState} from 'react';
import {createRoot} from 'react-dom/client';
import {
    useApp,
    useDocumentTheme,
    useHostStyleVariables,
} from '@modelcontextprotocol/ext-apps/react';
import {
    Area,
    Bar,
    CartesianGrid,
    ComposedChart,
    Line,
    Pie,
    PieChart,
    XAxis,
    YAxis,
} from 'recharts';
import {
    Check,
    ChevronDown,
    Copy,
    Download,
    Loader2,
    TriangleAlert,
} from 'lucide-react';
import {PrismLight as SyntaxHighlighter} from 'react-syntax-highlighter';
import sqlLang from 'react-syntax-highlighter/dist/esm/languages/prism/sql';
import oneDark from 'react-syntax-highlighter/dist/esm/styles/prism/one-dark';
import oneLight from 'react-syntax-highlighter/dist/esm/styles/prism/one-light';
import {Card} from '@/components/ui/card';
import {Button} from '@/components/ui/button';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
    ChartContainer,
    ChartLegend,
    ChartLegendContent,
    ChartTooltip,
    ChartTooltipContent,
} from '@/components/ui/chart';
import {
    toCartesianModel,
    toProportionalModel,
    type CartesianInput,
    type ProportionalInput,
    type Row,
} from './charts';

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

// ringRadii spreads concentric pie rings across the radial band (innermost
// first). A donut leaves a centre hole; a single-layer pie fills the centre.
const ringRadii = (
    index: number,
    total: number,
    donut: boolean,
): {inner: string; outer: string} => {
    const holeStart = donut ? 35 : 0;
    const end = 80;
    const width = (end - holeStart) / total;
    return {
        inner: `${holeStart + index * width}%`,
        outer: `${holeStart + (index + 1) * width}%`,
    };
};

// ChartView renders the active chart with shadcn/Recharts and the default
// --chart-N palette. cartesian -> ComposedChart (bar/line/area, stacking, flip);
// proportional -> PieChart (pie/donut, concentric layer rings).
const ChartView = ({
    chartKind,
    toolArgs,
    rows,
}: {
    chartKind: 'cartesian' | 'proportional';
    toolArgs: Record<string, unknown>;
    rows: Row[];
}) => {
    if (chartKind === 'cartesian') {
        const model = toCartesianModel(toolArgs as unknown as CartesianInput, rows);
        return (
            <ChartContainer config={model.config} className="min-h-72 w-full">
                <ComposedChart
                    data={model.data}
                    layout={model.flip ? 'vertical' : 'horizontal'}>
                    <CartesianGrid strokeDasharray="3 3" />
                    {model.flip ? (
                        <>
                            <XAxis type="number" tickLine={false} axisLine={false} />
                            <YAxis
                                type="category"
                                dataKey={model.xKey}
                                width={88}
                                tickLine={false}
                                axisLine={false}
                            />
                        </>
                    ) : (
                        <>
                            <XAxis
                                dataKey={model.xKey}
                                tickLine={false}
                                axisLine={false}
                            />
                            <YAxis tickLine={false} axisLine={false} />
                        </>
                    )}
                    <ChartTooltip content={<ChartTooltipContent />} />
                    <ChartLegend content={<ChartLegendContent />} />
                    {model.series.map((s) =>
                        s.kind === 'bar' ? (
                            <Bar
                                key={s.key}
                                dataKey={s.key}
                                name={s.label}
                                fill={s.color}
                                stackId={s.stackId}
                                radius={4}
                            />
                        ) : s.kind === 'area' ? (
                            <Area
                                key={s.key}
                                dataKey={s.key}
                                name={s.label}
                                type={s.type}
                                stroke={s.color}
                                fill={s.color}
                                fillOpacity={0.3}
                                stackId={s.stackId}
                                dot={false}
                            />
                        ) : (
                            <Line
                                key={s.key}
                                dataKey={s.key}
                                name={s.label}
                                type={s.type}
                                stroke={s.color}
                                strokeWidth={2}
                                dot={false}
                            />
                        ),
                    )}
                </ComposedChart>
            </ChartContainer>
        );
    }

    const model = toProportionalModel(toolArgs as unknown as ProportionalInput, rows);
    return (
        <ChartContainer config={model.config} className="min-h-72 w-full">
            <PieChart>
                <ChartTooltip
                    content={
                        <ChartTooltipContent
                            // Radial charts have no axis label, so the custom
                            // formatter (which replaces the whole row) must carry
                            // both the slice name (category) and the value.
                            formatter={(value, name) => (
                                <span className="flex w-full items-center justify-between gap-3">
                                    <span className="text-muted-foreground">
                                        {String(name)}
                                    </span>
                                    <span className="font-medium tabular-nums text-foreground">
                                        {model.format(Number(value))}
                                    </span>
                                </span>
                            )}
                        />
                    }
                />
                {model.layers.map((layer, li) => {
                    const {inner, outer} = ringRadii(
                        li,
                        model.layers.length,
                        model.donut,
                    );
                    return (
                        <Pie
                            key={li}
                            // Recharts v3 colors each slice from its datum's
                            // `fill`, so carry the palette color on the data.
                            data={layer.data.map((slice) => ({
                                ...slice,
                                fill: slice.color,
                            }))}
                            dataKey="value"
                            nameKey="name"
                            innerRadius={inner}
                            outerRadius={outer}
                            strokeWidth={1}
                        />
                    );
                })}
                <ChartLegend content={<ChartLegendContent nameKey="name" />} />
            </PieChart>
        </ChartContainer>
    );
};

const App = () => {
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
    const [copied, setCopied] = useState(false);

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
    const theme = useDocumentTheme();

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

    const copySql = async () => {
        const text = data?.sql ?? '';
        // The sandboxed host iframe often blocks navigator.clipboard, so fall back
        // to a hidden <textarea> + execCommand('copy') under the click gesture.
        try {
            await navigator.clipboard.writeText(text);
        } catch {
            const textarea = document.createElement('textarea');
            textarea.value = text;
            textarea.style.position = 'fixed';
            textarea.style.opacity = '0';
            document.body.appendChild(textarea);
            textarea.select();
            try {
                document.execCommand('copy');
            } finally {
                document.body.removeChild(textarea);
            }
        }
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
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
                        <DropdownMenu>
                            <DropdownMenuTrigger
                                render={
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        disabled={exporting !== null}
                                    />
                                }>
                                <Download size={14} />
                                Download
                                <ChevronDown size={14} />
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                                <DropdownMenuItem
                                    onClick={() => void exportAs('csv')}>
                                    CSV
                                </DropdownMenuItem>
                                <DropdownMenuItem
                                    onClick={() => void exportAs('json')}>
                                    JSON
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    )}
                </div>

                <TabsContent value="chart">
                    {chartKind ? (
                        <ChartView
                            chartKind={chartKind}
                            toolArgs={toolArgs ?? {}}
                            rows={data.rows}
                        />
                    ) : (
                        <p className="text-sm text-muted-foreground">
                            No chart mapping provided.
                        </p>
                    )}
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
                    <div className="relative">
                        {/* Floating copy button: sibling of the scroll area so it
                            stays pinned while the SQL scrolls underneath. */}
                        <Button
                            variant="outline"
                            size="icon"
                            className="absolute right-2 top-2 z-10 size-7"
                            aria-label="Copy SQL"
                            onClick={() => void copySql()}>
                            {copied ? <Check size={14} /> : <Copy size={14} />}
                        </Button>
                        <div className="max-h-72 overflow-auto rounded-md border border-border bg-muted text-sm">
                            <SyntaxHighlighter
                                language="sql"
                                style={theme === 'dark' ? oneDark : oneLight}
                                // The prism theme paints a background on both <pre>
                                // and <code>; clear both so the single bg-muted
                                // panel shows through uniformly, not per-line boxes.
                                customStyle={{
                                    margin: 0,
                                    padding: '0.75rem',
                                    background: 'transparent',
                                }}
                                codeTagProps={{
                                    style: {background: 'transparent'},
                                }}
                                wrapLongLines>
                                {data.sql || '-- no SQL available'}
                            </SyntaxHighlighter>
                        </div>
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
