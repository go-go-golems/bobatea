# Kubernetes Manifests Directory Diff (Example)

## Overview
This example compares two directories of Kubernetes YAML manifests and renders a diff per file. It’s useful for auditing changes between branches or environments (e.g., `staging` vs `prod`).

## What it demonstrates
- YAML decoding and flattening into path-value pairs
- One `DiffItem` per manifest file for quick navigation
- Redaction of common sensitive fields (`data.password`, `data.apiKey`, env `API_KEY`)
- Search across filenames, paths, and values

## How it’s built
- Walks both directories, collecting `*.yaml`/`*.yml` files
- Flattens each manifest using `internal/xdiff` and computes diffs
- Feeds a list of `DiffItem`s to the diff model

## Run
```bash
cd bobatea
GOWORK=off go run ./examples/diff/k8s-dir --before ./examples/diff/k8s-dir/before --after ./examples/diff/k8s-dir/after
```

## Keys
- Up/Down: switch file items
- Tab: focus detail pane
- `/`: search
- `r`: redact
- `q`: quit

## Sample Data
Create simple fixtures before running:
```bash
mkdir -p examples/diff/k8s-dir/before examples/diff/k8s-dir/after
cat > examples/diff/k8s-dir/before/deployment.yaml <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: svc
spec:
  replicas: 1
YAML
cat > examples/diff/k8s-dir/after/deployment.yaml <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: svc
spec:
  replicas: 2
YAML
```

## Notes
- For very large trees, consider implementing pagination or virtualization; see roadmap suggestions below.
