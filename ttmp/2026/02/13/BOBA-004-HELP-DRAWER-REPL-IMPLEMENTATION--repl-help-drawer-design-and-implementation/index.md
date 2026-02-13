---
Title: REPL Help Drawer Design and Implementation
Ticket: BOBA-004-HELP-DRAWER-REPL-IMPLEMENTATION
Status: active
Topics:
    - repl
    - help-drawer
    - implementation
    - analysis
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: bobatea/pkg/overlay/overlay.go
      Note: Initial overlay rendering mechanism for drawer
    - Path: bobatea/pkg/repl/model.go
      Note: Main code integration point for drawer lifecycle and updates
    - Path: bobatea/ttmp/2026/02/13/BOBA-004-HELP-DRAWER-REPL-IMPLEMENTATION--repl-help-drawer-design-and-implementation/design-doc/01-help-drawer-analysis-and-implementation-guide.md
      Note: Detailed help drawer analysis and implementation plan
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-13T09:59:55.40956153-05:00
WhatFor: ""
WhenToUse: ""
---


# REPL Help Drawer Design and Implementation

## Overview

This ticket captures the design and implementation path for a keyboard-toggle help drawer that can stay open and adapt live to typing/cursor context.
The design is overlay-layer friendly and remains compatible with replacing/rebuilding autocomplete components.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Primary Design Doc**: [design-doc/01-help-drawer-analysis-and-implementation-guide.md](./design-doc/01-help-drawer-analysis-and-implementation-guide.md)

## Status

Current status: **active**

## Topics

- repl
- help-drawer
- implementation
- analysis

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design-doc/ - Architecture and implementation design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
