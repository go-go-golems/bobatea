# JSON Directory Diff (Example)

## Overview
Compare two directories of JSON files and present one diff item per file. Each file is flattened and diffed independently; the list pane lets you jump between files.

## What it demonstrates
- Multi-item provider: one `DiffItem` per file
- Stable ordering and quick navigation across many items
- Redaction for common secret paths
- Useful for config repos and environment overlays

## How it’s built
- Recursively collects `*.json` files from `--before` and `--after` directories
- Flattens each file and computes a per-file diff
- Creates a `DiffItem` for each file present in either tree

## Run
```bash
cd bobatea
GOWORK=off go run ./examples/diff/json-dir --before ./examples/diff/json-dir/before --after ./examples/diff/json-dir/after
```

## Keys
- Up/Down: navigate files
- Tab: focus detail pane
- `/`: search across file names, paths, and values
- `r`: toggle redaction
- `q`: quit

## Sample Data
Create simple fixtures before running:
```bash
mkdir -p examples/diff/json-dir/before examples/diff/json-dir/after
cat > examples/diff/json-dir/before/app.json <<'JSON'
{ "name": "svc", "replicas": 1, "env": { "LOG_LEVEL": "info" } }
JSON
cat > examples/diff/json-dir/after/app.json <<'JSON'
{ "name": "svc", "replicas": 2, "env": { "LOG_LEVEL": "debug" } }
JSON
```
