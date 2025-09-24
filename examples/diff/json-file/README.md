# JSON File Diff (Example)

## Overview
This example shows how to build a minimal adapter from two JSON files into the `pkg/diff` component. It flattens nested objects into dot paths, computes a simple diff (added/removed/updated), and demonstrates search and redaction.

## What it demonstrates
- Minimal `DataProvider` adapter around file-based inputs
- Flattening deeply nested data into renderable change lines
- Sensitive key redaction (toggle with `r`)
- Substring search across paths and values (`/`)
- Two-pane navigation and layout sizing

## How it’s built
- `internal/xdiff` provides helpers to decode JSON/YAML and flatten into `map[string]any` keyed by dot/bracket paths.
- We compute a naive diff of flattened maps to produce `Change` lists.
- An adapter implements `DataProvider`, `DiffItem`, `Category`, and `Change` using simple structs.
- The diff model from `pkg/diff` is instantiated and run with Bubble Tea.

## Run
```bash
# From repository root
cd bobatea
GOWORK=off go run ./examples/diff/json-file --before ./examples/diff/json-file/before.json --after ./examples/diff/json-file/after.json
```

## Keys
- Up/Down: move in list
- Tab: switch pane
- `/`: search
- `r`: toggle redaction
- `q`: quit

## Why it’s interesting
- Portable: you can adapt any structured input to the minimal diff interfaces.
- Clear UX: two-pane layout highlights context and detail with very little code.
- Safe by default: example redacts common secrets (`password`, `api_key`, `secrets.token`).
