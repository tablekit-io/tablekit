import {createHash} from 'node:crypto';
import {existsSync, readFileSync, writeFileSync, rmSync} from 'node:fs';
import {fileURLToPath} from 'node:url';
import {join, relative, resolve} from 'node:path';
import {defineConfig, type Plugin} from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import {viteSingleFile} from 'vite-plugin-singlefile';

const ROOT = fileURLToPath(new URL('.', import.meta.url));

// One entry per MCP-app template. The key is the template name (also the
// manifest key + output filename stem); the value is its HTML entry. Add a
// template by dropping a folder under src/ and listing it here, then add the
// name to build.mjs's ENTRIES list.
const ALL_ENTRIES: Readonly<Record<string, string>> = {
    'chart_renderer': resolve(ROOT, 'src/chart_renderer/index.html'),
};

// viteSingleFile inlines everything into one chunk (codeSplitting:false), which
// rollup rejects with multiple inputs. So each template is built in its own
// single-input vite invocation, selected by WIDGET_ENTRY (see build.mjs); a
// plain `vite build` with no selection builds the whole set in one chunk only if
// there's a single template. The manifest is merged across invocations.
const selected = process.env.WIDGET_ENTRY;
const ENTRIES: Readonly<Record<string, string>> = selected
    ? {[selected]: ALL_ENTRIES[selected]}
    : ALL_ENTRIES;

const MiB = 1024 * 1024;
const WARN_BYTES = 16 * MiB;
const MAX_BYTES = 32 * MiB; // exactly 32 MiB — the host's ui:// transport ceiling

// After viteSingleFile has inlined every asset into one HTML per entry, this
// plugin finalises the build artifacts core consumes: it content-hashes each
// HTML, renames it to `<name>-<hash>.html`, enforces the size budget, and merges
// the template into manifest.json (template name -> {file, hash, bytes}). The
// hash is the cache key core advertises (ui://tablekit/<name>-<hash>), so any
// content change auto-busts the host's per-URI resource cache. Merging (rather
// than overwriting) lets each per-entry invocation contribute its own line.
const singlefileManifest = (
    entries: Readonly<Record<string, string>>,
): Plugin => {
    let outDir = '';
    let root = '';
    return {
        name: 'tablekit-singlefile-manifest',
        configResolved(config) {
            outDir = config.build.outDir;
            root = config.root;
        },
        closeBundle() {
            const manifestPath = join(outDir, 'manifest.json');
            const manifest: Record<
                string,
                {file: string; hash: string; bytes: number}
            > = existsSync(manifestPath)
                ? JSON.parse(readFileSync(manifestPath, 'utf8'))
                : {};

            for (const [name, input] of Object.entries(entries)) {
                // viteSingleFile emits the inlined HTML at the input's path
                // relative to root, under outDir.
                const emitted = join(outDir, relative(root, input));
                const bytes = readFileSync(emitted);
                if (bytes.length > MAX_BYTES) {
                    this.error(
                        `[singlefile] "${name}" is ${bytes.length} bytes (> 32 MiB hard cap) — refusing to emit`,
                    );
                }
                if (bytes.length > WARN_BYTES) {
                    this.warn(
                        `[singlefile] "${name}" is ${(bytes.length / MiB).toFixed(1)} MiB (> 16 MiB warning threshold)`,
                    );
                }
                const hash = createHash('sha256')
                    .update(bytes)
                    .digest('hex')
                    .slice(0, 16);
                const file = `${name}-${hash}.html`;
                writeFileSync(join(outDir, file), bytes);
                rmSync(emitted);
                manifest[name] = {file, hash, bytes: bytes.length};
            }
            // Drop the now-empty nested input dirs left under outDir.
            rmSync(join(outDir, 'src'), {recursive: true, force: true});
            writeFileSync(
                manifestPath,
                JSON.stringify(manifest, null, 4) + '\n',
            );
        },
    };
};

export default defineConfig({
    root: ROOT,
    // __WIDGET_DEV__ is true only for the watch (dev) build, so widget code can
    // gate dev-only behaviour (verbose console logging) and have it dead-code
    // eliminated from production builds. Set by build.mjs.
    define: {
        __WIDGET_DEV__: JSON.stringify(process.env.WIDGET_DEV === 'true'),
    },
    plugins: [
        tailwindcss(),
        react(),
        viteSingleFile(),
        singlefileManifest(ENTRIES),
    ],
    resolve: {
        // @ resolves shadcn's "@/..." imports to src.
        alias: {
            '@': resolve(ROOT, 'src'),
        },
    },
    build: {
        outDir: resolve(ROOT, 'dist'),
        // build.mjs cleans dist once up front; per-entry invocations must not wipe
        // each other's output (or the merged manifest).
        emptyOutDir: false,
        rollupOptions: {input: ENTRIES},
    },
});
