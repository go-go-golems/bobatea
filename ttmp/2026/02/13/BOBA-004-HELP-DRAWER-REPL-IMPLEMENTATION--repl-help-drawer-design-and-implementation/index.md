---
Title: REPL Help Drawer Design and Implementation
Ticket: BOBA-004-HELP-DRAWER-REPL-IMPLEMENTATION
Status: complete
Topics:
    - repl
    - help-drawer
    - implementation
    - analysis
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/repl/help_drawer_types.go
      Note: Help drawer provider contracts
    - Path: pkg/repl/config.go
      Note: Drawer config defaults and tuning
    - Path: pkg/repl/keymap.go
      Note: Drawer key bindings
    - Path: pkg/repl/model.go
      Note: Drawer lifecycle, debounce, and overlay rendering integration
    - Path: pkg/repl/help_drawer_model_test.go
      Note: Coverage for toggle, adaptive updates, and stale filtering
    - Path: examples/repl/autocomplete-generic/main.go
      Note: Manual demo provider for drawer behavior
    - Path: ttmp/2026/02/13/BOBA-004-HELP-DRAWER-REPL-IMPLEMENTATION--repl-help-drawer-design-and-implementation/design-doc/01-help-drawer-analysis-and-implementation-guide.md
      Note: Detailed help drawer analysis and implementation plan
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-13T18:01:21.146375918-05:00
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

Current status: **complete**

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
