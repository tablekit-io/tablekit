// Build orchestrator: each MCP-app template is built in its own single-input
// vite invocation, because vite-plugin-singlefile inlines everything into one
// chunk (codeSplitting:false), which rollup rejects when there are multiple
// inputs. We clean dist once, then build each template with WIDGET_ENTRY set;
// the manifest plugin merges each template into manifest.json. With --watch,
// one watcher per template runs concurrently.
import {spawn} from 'node:child_process';
import {mkdirSync, readdirSync, rmSync} from 'node:fs';
import {dirname, join, resolve} from 'node:path';
import {fileURLToPath} from 'node:url';

const ROOT = dirname(fileURLToPath(import.meta.url));

// The templates to build, one per src/<name>/ folder. Keep in sync with
// vite.config.ts's ALL_ENTRIES.
const ENTRIES = ['hello_world_interactive', 'chart_renderer'];

const watch = process.argv.includes('--watch');
const viteBin = resolve(ROOT, 'node_modules/.bin/vite');

// Clean dist's contents once (not the dir itself: on macOS the provenance xattr
// can block removing a directory created by a sandboxed tool, though removing
// its entries is allowed). Per-entry builds then append and merge the manifest.
const distDir = resolve(ROOT, 'dist');
mkdirSync(distDir, {recursive: true});
for (const entry of readdirSync(distDir)) {
    rmSync(join(distDir, entry), {recursive: true, force: true});
}

const runEntry = (name) =>
    new Promise((resolveEntry, rejectEntry) => {
        const args = watch ? ['build', '--watch'] : ['build'];
        const child = spawn(viteBin, args, {
            cwd: ROOT,
            stdio: 'inherit',
            env: {...process.env, WIDGET_ENTRY: name},
        });
        child.on('exit', (code) =>
            code === 0
                ? resolveEntry()
                : rejectEntry(new Error(`build "${name}" exited with ${code}`)),
        );
    });

if (watch) {
    // Concurrent watchers; they run until interrupted.
    ENTRIES.forEach((name) => void runEntry(name));
} else {
    // Sequential so manifest merges are race-free.
    for (const name of ENTRIES) {
        await runEntry(name);
    }
}
