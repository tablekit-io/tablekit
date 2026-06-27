import {createHash} from 'node:crypto';
import {readFileSync, writeFileSync, rmSync} from 'node:fs';
import {fileURLToPath} from 'node:url';
import {join, relative, resolve} from 'node:path';
import {defineConfig, type Plugin} from 'vite';
import preact from '@preact/preset-vite';
import {viteSingleFile} from 'vite-plugin-singlefile';

const ROOT = fileURLToPath(new URL('.', import.meta.url));

// One entry per MCP-app template. The key is the template name (also the
// manifest key + output filename stem); the value is its HTML entry. Add a
// template by dropping a folder under src/ and listing it here — nothing
// else in the pipeline needs to change.
const ENTRIES: Readonly<Record<string, string>> = {
    'hello_world_interactive': resolve(
        ROOT,
        'src/hello_world_interactive/index.html',
    ),
};

const MiB = 1024 * 1024;
const WARN_BYTES = 16 * MiB;
const MAX_BYTES = 32 * MiB; // exactly 32 MiB — the host's ui:// transport ceiling

// After viteSingleFile has inlined every asset into one HTML per entry, this
// plugin finalises the build artifacts core consumes: it content-hashes each
// HTML, renames it to `<name>-<hash>.html`, enforces the size budget, and
// writes a manifest mapping template name -> {file, hash, bytes}. The hash is
// the cache key core advertises (ui://tablekit/<name>-<hash>), so any content
// change auto-busts the host's per-URI resource cache.
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
            const built = Object.entries(entries).map(([name, input]) => {
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
                return [name, {file, hash, bytes: bytes.length}] as const;
            });
            // Drop the now-empty nested input dirs left under outDir.
            rmSync(join(outDir, 'src'), {recursive: true, force: true});
            writeFileSync(
                join(outDir, 'manifest.json'),
                JSON.stringify(Object.fromEntries(built), null, 4) + '\n',
            );
        },
    };
};

export default defineConfig({
    root: ROOT,
    plugins: [preact(), viteSingleFile(), singlefileManifest(ENTRIES)],
    build: {
        outDir: resolve(ROOT, 'dist'),
        emptyOutDir: true,
        rollupOptions: {input: ENTRIES},
    },
});
