// Pure, framework-agnostic builders that shape fetch_chart_data rows + the render
// tool's mapping into Recharts-ready models. Kept free of JSX/DOM so the mapping
// logic is easy to read; the Chart tab in main.tsx renders these with shadcn's
// chart components. Colors use the default shadcn chart tokens (--chart-1..5),
// theme-aware for free; an explicit color_hue overrides.
import {type ChartConfig} from '@/components/ui/chart';

// Row is one result row keyed by column name, as fetch_chart_data returns it.
export type Row = Record<string, unknown>;

// CartesianInput is show_bar_line_area_chart's arguments.
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

// ProportionalInput is show_pie_donut_sunburst_chart's arguments.
export type ProportionalInput = {
    readonly query_key: string;
    readonly display?: 'pie' | 'donut';
    readonly value_prop: string;
    readonly value_prefix?: string;
    readonly value_suffix?: string;
    readonly layers: ReadonlyArray<{readonly discriminator_prop: string}>;
};

// num coerces a normalized cell value to a number; non-numeric/blank cells count
// as 0 so a stray null doesn't break a series.
const num = (value: unknown): number => {
    const n = typeof value === 'number' ? value : Number(value);
    return Number.isFinite(n) ? n : 0;
};

// colorAt picks a slice/series color: the shadcn default palette cycled over
// --chart-1..5, or an explicit hue when the render tool supplied a valid one.
// The hue is coerced to a finite 0–359 number so a malformed value can never
// produce a color string that breaks out of the generated <style> (see the
// SAFE_COLOR guard in components/ui/chart.tsx).
const colorAt = (index: number, hue?: number): string => {
    const h = Number(hue);
    if (Number.isFinite(h)) {
        return `hsl(${((h % 360) + 360) % 360} 65% 50%)`;
    }
    return `var(--chart-${(index % 5) + 1})`;
};

// CartesianSeries describes one Y series for the ComposedChart.
export type CartesianSeries = {
    readonly key: string;
    readonly label: string;
    readonly kind: 'bar' | 'line' | 'area';
    readonly type: 'monotone' | 'linear';
    readonly stackId?: string;
    readonly color: string;
};

export type CartesianModel = {
    readonly data: Row[];
    readonly series: CartesianSeries[];
    readonly config: ChartConfig;
    readonly xKey: string;
    readonly xLabel: string;
    readonly flip: boolean;
};

// toCartesianModel builds the data + series + ChartConfig a ComposedChart needs.
// Y columns are coerced to numbers; display_as -> kind, shape -> curve/linear,
// stack_group -> stackId.
export function toCartesianModel(
    input: CartesianInput,
    rows: readonly Row[],
): CartesianModel {
    const yKeys = input.y.map((s) => s.prop);
    const data = rows.map((row) => {
        const copy: Row = {...row};
        for (const key of yKeys) {
            copy[key] = num(row[key]);
        }
        return copy;
    });

    const series: CartesianSeries[] = input.y.map((s, i) => ({
        key: s.prop,
        label: s.axes_label || s.prop,
        kind:
            s.display_as === 'line'
                ? 'line'
                : s.display_as === 'area'
                  ? 'area'
                  : 'bar',
        type: s.shape === 'curve' ? 'monotone' : 'linear',
        stackId: s.stack_group || undefined,
        color: colorAt(i, s.color_hue),
    }));

    const config: ChartConfig = Object.fromEntries(
        series.map((s) => [s.key, {label: s.label, color: s.color}]),
    );

    return {
        data,
        series,
        config,
        xKey: input.x.prop,
        xLabel: input.x.axes_label,
        flip: !!input.flip_axes,
    };
}

// ProportionalSlice is one slice of one ring.
export type ProportionalSlice = {
    readonly name: string;
    readonly value: number;
    readonly color: string;
};

export type ProportionalModel = {
    readonly layers: ReadonlyArray<{readonly data: ProportionalSlice[]}>;
    readonly donut: boolean;
    readonly config: ChartConfig;
    readonly format: (value: number) => string;
};

// toProportionalModel groups rows per layer (innermost first) by its column,
// summing value_prop, and builds a config + value formatter (prefix/suffix).
export function toProportionalModel(
    input: ProportionalInput,
    rows: readonly Row[],
): ProportionalModel {
    const layers = input.layers.map((layer) => {
        const groups = new Map<string, number>();
        for (const row of rows) {
            const key = String(row[layer.discriminator_prop] ?? '');
            groups.set(key, (groups.get(key) ?? 0) + num(row[input.value_prop]));
        }
        const keys = [...groups.keys()];
        return {
            data: keys.map((name, i) => ({
                name,
                value: groups.get(name) ?? 0,
                color: colorAt(i),
            })),
        };
    });

    const config: ChartConfig = {};
    for (const layer of layers) {
        for (const slice of layer.data) {
            config[slice.name] = {label: slice.name, color: slice.color};
        }
    }

    const format = (value: number): string =>
        `${input.value_prefix ?? ''}${value}${input.value_suffix ?? ''}`;

    return {layers, donut: input.display !== 'pie', config, format};
}
