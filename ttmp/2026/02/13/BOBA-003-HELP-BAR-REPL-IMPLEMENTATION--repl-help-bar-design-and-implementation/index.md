---
Title: REPL Help Bar Design and Implementation
Ticket: BOBA-003-HELP-BAR-REPL-IMPLEMENTATION
Status: active
Topics:
    - repl
    - help-bar
    - implementation
    - analysis
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: bobatea/pkg/repl/config.go
      Note: Configuration surface for new help bar feature flags
    - Path: bobatea/pkg/repl/model.go
      Note: Main code integration point for input scheduling and rendering
    - Path: bobatea/ttmp/2026/02/13/BOBA-003-HELP-BAR-REPL-IMPLEMENTATION--repl-help-bar-design-and-implementation/design-doc/01-help-bar-analysis-and-implementation-guide.md
      Note: Detailed help bar analysis and implementation plan
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-13T09:59:55.419361042-05:00
WhatFor: ""
WhenToUse: ""
---


# REPL Help Bar Design and Implementation

## Overview

This ticket captures the design and implementation path for a contextual REPL help bar that updates while typing (debounced), with semantic trigger/show decisions delegated to provider logic.
The architecture is intentionally compatible with replacing or rewriting the current autocomplete widget.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Primary Design Doc**: [design-doc/01-help-bar-analysis-and-implementation-guide.md](./design-doc/01-help-bar-analysis-and-implementation-guide.md)
- **Implementation Diary**: [reference/01-diary.md](./reference/01-diary.md)

## Status

Current status: **active**

## Topics

- repl
- help-bar
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
