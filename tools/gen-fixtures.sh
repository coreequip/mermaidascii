#!/usr/bin/env bash
# Erzeugt testdata/golden/ neu aus der Originalbibliothek.
#
# Die Golden-Dateien sind die Ausgabe von beautiful-mermaid, gegen die diese
# Portierung geprüft wird. Sie sind Referenz, nicht Wahrheit: weicht die
# Go-Fassung ab, ist zuerst zu klären, welche der beiden recht hat. Bewusst
# abweichende Stellen gehören nach docs/DEVIATIONS.md, nicht stillschweigend
# in die Golden-Dateien.
#
# Braucht bun und Netzzugriff. Nur ausführen, wenn Fixtures dazukommen oder
# die Referenzfassung angehoben wird.
set -euo pipefail

# Die Fassung, gegen die portiert wurde. Anheben ist eine bewusste Entscheidung.
COMMIT=2ac8bbbb060ca0a65a6a21f3200bd99b1587b488

repo=$(cd "$(dirname "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

git clone -q https://github.com/lukilabs/beautiful-mermaid.git "$work/bm"
git -C "$work/bm" checkout -q "$COMMIT"
(cd "$work/bm" && bun install >/dev/null)

cat > "$work/bm/gen.ts" <<'TS'
import { renderMermaidASCII } from './src/ascii/index.ts'
import { readdirSync, writeFileSync, mkdirSync } from 'node:fs'
import { join } from 'node:path'

const inDir = process.argv[2]
const outDir = process.argv[3]
mkdirSync(outDir, { recursive: true })

for (const f of readdirSync(inDir).filter((f) => f.endsWith('.mmd')).sort()) {
  const text = await Bun.file(join(inDir, f)).text()
  const base = f.replace(/\.mmd$/, '')
  for (const [suffix, useAscii] of [['unicode', false], ['ascii', true]] as const) {
    const out = renderMermaidASCII(text, { colorMode: 'none', useAscii })
    // Leerraum am Zeilenende weg: er ist im Original Zufall der Zeichenfläche
    // und würde die Golden-Vergleiche unnötig brüchig machen.
    writeFileSync(join(outDir, `${base}.${suffix}.txt`), out.replace(/[ \t]+$/gm, '') + '\n')
    console.log(`  ${base}.${suffix}`)
  }
}
TS

bun "$work/bm/gen.ts" "$repo/testdata/input" "$repo/testdata/golden"
echo "Golden-Dateien neu geschrieben aus beautiful-mermaid@${COMMIT:0:12}"
