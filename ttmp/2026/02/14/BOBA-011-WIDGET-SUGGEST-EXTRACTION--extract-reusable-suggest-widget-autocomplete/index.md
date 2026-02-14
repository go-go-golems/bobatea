---
Title: Extract Reusable Suggest Widget (Autocomplete)
Ticket: BOBA-011-WIDGET-SUGGEST-EXTRACTION
Status: active
Topics:
    - cleanup
    - repl
    - widgets
    - autocomplete
    - implementation
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/repl/autocomplete_model_test.go
      Note: Assertion updates and parity checks for extracted suggest path
    - Path: pkg/repl/autocomplete_types.go
      Note: Type aliases to suggest request/result/reason types
    - Path: pkg/repl/completion_model.go
      Note: REPL adapter and widget state synchronization
    - Path: pkg/repl/completion_overlay.go
      Note: Overlay layout/render delegation to suggest widget
    - Path: pkg/repl/model_async_provider.go
      Note: Completion command delegation through widget provider runner
    - Path: pkg/repl/model_messages.go
      Note: Message aliases for debounce/result/layout
    - Path: pkg/tui/widgets/suggest/types.go
      Note: Shared contracts for request/result/messages and overlay config
    - Path: pkg/tui/widgets/suggest/widget.go
      Note: Reusable suggest widget state machine and provider orchestration
    - Path: pkg/tui/widgets/suggest/overlay.go
      Note: Extracted geometry calculations for overlay placement and clamping
    - Path: pkg/tui/widgets/suggest/render.go
      Note: Popup rendering with style injection and no-border mode
    - Path: pkg/tui/widgets/suggest/widget_test.go
      Note: Widget-level behavior and geometry parity tests
ExternalSources: []
Summary: "Extracts REPL autocomplete behavior into reusable suggest widget and integrates via adapter while preserving UX parity."
LastUpdated: 2026-02-14T17:35:00-05:00
WhatFor: "Track BOBA-011 implementation status and review artifacts for suggest widget extraction."
WhenToUse: "Use when reviewing autocomplete extraction changes or validating behavior parity in examples."
---

# Extract Reusable Suggest Widget (Autocomplete)

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
- autocomplete
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
