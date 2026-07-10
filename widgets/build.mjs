// Build orchestrator: each MCP-app template is built in its own vite build,
// because vite-plugin-singlefile forces output.codeSplitting=false to inline
// everything into one chunk, which rollup rejects when there are multiple
// inputs. So we can't do one build with N inputs — but we can loop vite's
// in-process build() API over the templates (no subprocesses). WIDGET_ENTRY
// selects the single entry for each pass; the manifest plugin merges each into
// manifest.json. With --watch, every template gets its own in-process watcher.
import {build} from 'vite';
import {
    existsSync,
    mkdirSync,
    readdirSync,
    rmSync,
    writeFileSync,
} from 'node:fs';
import {dirname, join, resolve} from 'node:path';
import {fileURLToPath} from 'node:url';

const ROOT = dirname(fileURLToPath(import.meta.url));

// The templates to build, one per src/<name>/ folder. Keep in sync with
// vite.config.ts's ALL_ENTRIES.
const ENTRIES = ['chart_renderer'];

const watch = process.argv.includes('--watch');

// The watch build IS the dev flow (`bun run dev`); a one-shot build is a
// production artifact. Surface that to vite.config.ts so it can compile in (or
// out) the widget's verbose console logging via the __WIDGET_DEV__ define.
process.env.WIDGET_DEV = watch ? 'true' : 'false';

// Clean dist's contents once (not the dir itself: on macOS the provenance xattr
// can block removing a directory created by a sandboxed tool, though removing
// its entries is allowed). Per-template builds then append and merge the manifest.
//
// Preserve .gitkeep: in dev this dist/ is bind-mounted onto core/mcp/ui/widgets
// (the //go:embed all:widgets dir), where a committed .gitkeep keeps the embed
// compiling before the first build. Deleting it here would break `go build` on a
// clean tree, so we skip it and re-create it if it's missing.
const GITKEEP = '.gitkeep';
const distDir = resolve(ROOT, 'dist');
mkdirSync(distDir, {recursive: true});
for (const entry of readdirSync(distDir)) {
    if (entry === GITKEEP) {
        continue;
    }
    rmSync(join(distDir, entry), {recursive: true, force: true});
}
const gitkeepPath = join(distDir, GITKEEP);
if (!existsSync(gitkeepPath)) {
    writeFileSync(gitkeepPath, '');
}

// vite re-reads the config (a function of process.env) on each build() call, so
// setting WIDGET_ENTRY before each pass scopes that build to one template.
for (const name of ENTRIES) {
    process.env.WIDGET_ENTRY = name;
    await build(watch ? {build: {watch: {}}} : {});
}
