---
Title: Extract Reusable Context Bar Widget (Help Bar)
Ticket: BOBA-012-WIDGET-CONTEXTBAR-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - help-bar
    - implementation
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/repl/help_bar_model_test.go
      Note: Adapted REPL help bar tests for extracted widget path
    - Path: pkg/repl/help_bar_types.go
      Note: Repl help bar type aliases to extracted contextbar types
    - Path: pkg/repl/helpbar_model.go
      Note: REPL adapter wiring from HelpBarProvider to contextbar.Widget
    - Path: pkg/repl/model_messages.go
      Note: Message aliases for help bar debounce/result bridging
    - Path: pkg/tui/widgets/contextbar/types.go
      Note: Shared request/payload/reason contracts for context bar hosts
    - Path: pkg/tui/widgets/contextbar/widget.go
      Note: Reusable context bar state machine and async provider command orchestration
    - Path: pkg/tui/widgets/contextbar/widget_test.go
      Note: Widget-level debounce/stale/error/render test coverage
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-14T16:51:39.447811184-05:00
WhatFor: ""
WhenToUse: ""
---


# Extract Reusable Context Bar Widget (Help Bar)

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
- help-bar
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
