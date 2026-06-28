// Pure chart.js config builders for the chart_renderer widget. Kept free of DOM,
// the bridge and Preact so the mapping logic (rows + tool arguments -> a
// ChartConfiguration) is easy to read and unit-testable. The widget shell in
// main.tsx owns the canvas, the bridge handshake and the data fetch.
import {type ChartConfiguration, type ChartType} from 'chart.js';

// Row is one result row keyed by column name, as fetch_chart_data returns it.
export type Row = Record<string, unknown>;

// TTheme mirrors the host theme so chart text/grid colors read on both surfaces.
export type TTheme = 'light' | 'dark';

// CartesianInput is render_cartesian_series_chart's arguments (the axis/series
// mapping the host forwards as tool-input).
export type CartesianInput = {
    readonly query_key: string;
    readonly flip_axes?: boolean;
    readonly x: {readonly prop: string; readonly axes_label: string};
    readonly y: ReadonlyArray<{
        readonly prop: string;
        readonly axes_label: string;
        readonly display_as?: 'line' | 'area' | 'bar';
        readonly shape?: 'line' | 'discrete' | 'curve';
        readonly color_hue?: number;
        readonly stack_group?: string;
    }>;
};

// ProportionalInput is render_proportional_chart's arguments.
export type ProportionalInput = {
    readonly query_key: string;
    readonly display?: 'pie' | 'donut';
    readonly value_prop: string;
    readonly value_prefix?: string;
    readonly value_suffix?: string;
    readonly layers: ReadonlyArray<{readonly discriminator_prop: string}>;
};

// num coerces a normalized cell value to a number; non-numeric/blank cells count
// as 0 so a stray null doesn't blow up the whole series.
const num = (value: unknown): number => {
    const n = typeof value === 'number' ? value : Number(value);
    return Number.isFinite(n) ? n : 0;
};

// hsl renders an HSL color, optionally with alpha (modern slash syntax, which
// canvas accepts).
const hsl = (hue: number, sat: number, light: number, alpha = 1): string =>
    alpha >= 1
        ? `hsl(${hue} ${sat}% ${light}%)`
        : `hsl(${hue} ${sat}% ${light}% / ${alpha})`;

// colorFor picks a series/slice color. An explicit hue (render arg) wins;
// otherwise hues are spread by the golden angle so adjacent series stay
// distinct, with saturation/lightness tuned per theme.
const colorFor = (
    index: number,
    theme: TTheme,
    hue?: number,
    alpha = 1,
): string => {
    const h = hue ?? (index * 137.508) % 360;
    const sat = theme === 'dark' ? 65 : 62;
    const light = theme === 'dark' ? 60 : 48;
    return hsl(h, sat, light, alpha);
};

// themeColors are the axis text and grid-line colors for a theme.
const themeColors = (theme: TTheme) => ({
    text: theme === 'dark' ? '#e5e7eb' : '#374151',
    grid: theme === 'dark' ? 'rgba(255,255,255,0.10)' : 'rgba(0,0,0,0.10)',
});

// buildCartesianConfig maps rows + a cartesian mapping to a chart.js config. Each
// Y series becomes a dataset; bars and lines/areas can mix in one chart, series
// sharing a stack_group stack together, and flip_axes draws horizontally.
export function buildCartesianConfig(
    input: CartesianInput,
    rows: readonly Row[],
    theme: TTheme,
): ChartConfiguration {
    const colors = themeColors(theme);
    const labels = rows.map((row) => String(row[input.x.prop] ?? ''));
    const anyStack = input.y.some((series) => !!series.stack_group);

    const datasets = input.y.map((series, i) => {
        const isArea = series.display_as === 'area';
        const isLine = series.display_as === 'line' || isArea;
        const color = colorFor(i, theme, series.color_hue);
        return {
            type: (isLine ? 'line' : 'bar') as ChartType,
            label: series.axes_label || series.prop,
            data: rows.map((row) => num(row[series.prop])),
            backgroundColor: isArea
                ? colorFor(i, theme, series.color_hue, 0.25)
                : color,
            borderColor: color,
            borderWidth: isLine ? 2 : 0,
            fill: isArea,
            tension: series.shape === 'curve' ? 0.4 : 0,
            pointRadius: isLine ? 2 : 0,
            stack: series.stack_group || undefined,
        };
    });

    const valueScale = {
        ticks: {color: colors.text},
        grid: {color: colors.grid},
        stacked: anyStack,
    };
    const categoryScale = {
        title: {
            display: !!input.x.axes_label,
            text: input.x.axes_label,
            color: colors.text,
        },
        ticks: {color: colors.text},
        grid: {color: colors.grid},
        stacked: anyStack,
    };

    return {
        type: 'bar',
        data: {labels, datasets},
        options: {
            responsive: true,
            maintainAspectRatio: false,
            indexAxis: input.flip_axes ? 'y' : 'x',
            scales: input.flip_axes
                ? {x: valueScale, y: categoryScale}
                : {x: categoryScale, y: valueScale},
            plugins: {legend: {labels: {color: colors.text}}},
        },
    };
}

// buildProportionalConfig maps rows + a proportional mapping to a pie/donut
// config. Each layer groups rows by its column and sums value_prop, becoming one
// dataset; multiple layers render as concentric rings (innermost first).
export function buildProportionalConfig(
    input: ProportionalInput,
    rows: readonly Row[],
    theme: TTheme,
): ChartConfiguration {
    const colors = themeColors(theme);

    const datasets = input.layers.map((layer) => {
        const groups = new Map<string, number>();
        for (const row of rows) {
            const key = String(row[layer.discriminator_prop] ?? '');
            groups.set(key, (groups.get(key) ?? 0) + num(row[input.value_prop]));
        }
        const keys = [...groups.keys()];
        return {
            label: layer.discriminator_prop,
            data: keys.map((key) => groups.get(key) ?? 0),
            backgroundColor: keys.map((_, i) => colorFor(i, theme)),
            borderWidth: 1,
            // Each ring has its own slice keys; stash them so the tooltip can
            // name the right slice per dataset (the labels array only covers the
            // innermost ring's legend).
            keys,
        };
    });

    const fmt = (value: number): string =>
        `${input.value_prefix ?? ''}${value}${input.value_suffix ?? ''}`;

    return {
        type: input.display === 'pie' ? 'pie' : 'doughnut',
        data: {labels: datasets[0]?.keys ?? [], datasets},
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {position: 'right', labels: {color: colors.text}},
                tooltip: {
                    callbacks: {
                        label: (ctx) => {
                            const ds = ctx.dataset as {keys?: string[]};
                            const key =
                                ds.keys?.[ctx.dataIndex] ?? String(ctx.label);
                            return `${key}: ${fmt(num(ctx.parsed))}`;
                        },
                    },
                },
            },
        },
    };
}
