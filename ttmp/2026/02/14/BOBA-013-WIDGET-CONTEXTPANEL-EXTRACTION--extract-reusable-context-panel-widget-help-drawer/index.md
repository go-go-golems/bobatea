---
Title: Extract Reusable Context Panel Widget (Help Drawer)
Ticket: BOBA-013-WIDGET-CONTEXTPANEL-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - help-drawer
    - implementation
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/repl/help_drawer_types.go
      Note: Help drawer contract aliases to contextpanel types
    - Path: pkg/repl/helpdrawer_model.go
      Note: Repl adapter to contextpanel widget with legacy field sync
    - Path: pkg/repl/helpdrawer_overlay.go
      Note: Delegates overlay layout/render to contextpanel widget
    - Path: pkg/tui/widgets/contextpanel/overlay.go
      Note: Docked overlay layout computation and clamp behavior
    - Path: pkg/tui/widgets/contextpanel/render.go
      Note: Panel rendering composition for title/subtitle/body/footer
    - Path: pkg/tui/widgets/contextpanel/widget.go
      Note: Reusable context panel lifecycle and provider orchestration
    - Path: pkg/tui/widgets/contextpanel/widget_test.go
      Note: Widget unit coverage for toggle/pin/debounce/stale/layout/render
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-14T16:51:39.606806729-05:00
WhatFor: ""
WhenToUse: ""
---


# Extract Reusable Context Panel Widget (Help Drawer)

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- cleanup
- repl
- widgets
- help-drawer
- implementation

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
