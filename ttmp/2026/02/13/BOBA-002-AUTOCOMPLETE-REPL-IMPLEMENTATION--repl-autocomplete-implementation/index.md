---
Title: REPL Autocomplete Implementation
Ticket: BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION
Status: complete
Topics:
    - repl
    - autocomplete
    - implementation
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: bobatea/pkg/repl/config.go
      Note: Autocomplete and keybinding configuration surface
    - Path: bobatea/pkg/repl/evaluator.go
      Note: Optional completer capability contract extension point
    - Path: bobatea/pkg/repl/model.go
      Note: Main location for key routing
    - Path: bobatea/ttmp/2026/02/13/BOBA-002-AUTOCOMPLETE-REPL-IMPLEMENTATION--repl-autocomplete-implementation/design-doc/01-autocomplete-implementation-guide.md
      Note: Detailed implementation blueprint for debounce-driven autocomplete integration
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-13T15:39:36.191703135-05:00
WhatFor: ""
WhenToUse: ""
---



# REPL Autocomplete Implementation

## Overview

This ticket captures implementation-ready design guidance for integrating autocomplete into `bobatea/pkg/repl` with strict trigger-boundary rules:

- REPL handles debounce scheduling and explicit shortcut dispatch.
- Input completer decides whether completion should trigger/show.
- `Tab` can be used as an explicit trigger if configured; focus-toggle behavior is remapped accordingly.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Primary Design Doc**: [design-doc/01-autocomplete-implementation-guide.md](./design-doc/01-autocomplete-implementation-guide.md)

## Status

Current status: **active**

## Topics

- repl
- autocomplete
- implementation

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
