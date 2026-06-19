#!/usr/bin/env bash
set -euo pipefail

version="$(awk '/^version:/ { print $2; exit }' pubspec.yaml)"
package="singbox_ffi-${version}"

mkdir -p dist "${package}"
cp .pubignore .pubignore.pubdev

python3 - <<'PY'
from pathlib import Path

Path(".pubignore").write_text(
    "\n".join(
        [
            ".agents/",
            ".codex/",
            ".github/",
            ".gocache/",
            "downloaded-artifacts/",
            "examples/",
            "build/",
            "dist/",
            "singbox_ffi-*/",
            "cache.db",
            "configuration.json",
            "CrashReport-*.log",
            "*.exe",
            "*.test",
            "*.out",
            "",
        ]
    ),
    encoding="utf-8",
)
PY

rsync -a \
  --exclude .git \
  --exclude downloaded-artifacts \
  --exclude dist \
  --exclude "${package}" \
  ./ "${package}/"

mv .pubignore.pubdev .pubignore
zip -r "dist/${package}.zip" "${package}"
tar -czf "dist/${package}.tar.gz" "${package}"
rm -rf "${package}" downloaded-artifacts
