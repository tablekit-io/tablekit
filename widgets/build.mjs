// Build orchestrator: each MCP-app template is built in its own vite build,
// because vite-plugin-singlefile forces output.codeSplitting=false to inline
// everything into one chunk, which rollup rejects when there are multiple
// inputs. So we can't do one build with N inputs — but we can loop vite's
// in-process build() API over the templates (no subprocesses). WIDGET_ENTRY
// selects the single entry for each pass; the manifest plugin merges each into
// manifest.json. With --watch, every template gets its own in-process watcher.
import {build} from 'vite';
import {mkdirSync, readdirSync, rmSync} from 'node:fs';
import {dirname, join, resolve} from 'node:path';
import {fileURLToPath} from 'node:url';

const ROOT = dirname(fileURLToPath(import.meta.url));

// The templates to build, one per src/<name>/ folder. Keep in sync with
// vite.config.ts's ALL_ENTRIES.
const ENTRIES = ['chart_renderer'];

const watch = process.argv.includes('--watch');

// Clean dist's contents once (not the dir itself: on macOS the provenance xattr
// can block removing a directory created by a sandboxed tool, though removing
// its entries is allowed). Per-template builds then append and merge the manifest.
const distDir = resolve(ROOT, 'dist');
mkdirSync(distDir, {recursive: true});
for (const entry of readdirSync(distDir)) {
    rmSync(join(distDir, entry), {recursive: true, force: true});
}

// vite re-reads the config (a function of process.env) on each build() call, so
// setting WIDGET_ENTRY before each pass scopes that build to one template.
for (const name of ENTRIES) {
    process.env.WIDGET_ENTRY = name;
    await build(watch ? {build: {watch: {}}} : {});
}
