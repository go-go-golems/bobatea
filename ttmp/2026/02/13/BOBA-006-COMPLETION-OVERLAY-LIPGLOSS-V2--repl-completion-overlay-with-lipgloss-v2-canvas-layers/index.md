---
Title: REPL Completion Overlay with Lipgloss v2 Canvas Layers
Ticket: BOBA-006-COMPLETION-OVERLAY-LIPGLOSS-V2
Status: complete
Topics:
    - analysis
    - repl
    - autocomplete
    - lipgloss
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-02-13/integrate-ast-parser-repl/bobatea/pkg/repl/model.go
      Note: Current inline completion rendering and update flow
    - Path: /home/manuel/code/wesen/corporate-headquarters/2026-02-10--grail-js/internal/grailui/view.go
      Note: Lipgloss v2 layered rendering reference
ExternalSources:
    - https://github.com/charmbracelet/lipgloss/discussions/506
Summary: Analyze and design migration of REPL completion popup from inline rendering to lipgloss v2 overlay layers with size constraints and scrolling/paging behavior.
LastUpdated: 2026-02-13T19:09:54.891084531-05:00
WhatFor: Track BOBA-006 analysis artifacts and implementation guidance for overlay-based completion rendering.
WhenToUse: Use when implementing or reviewing REPL completion overlay work.
---


# REPL Completion Overlay with Lipgloss v2 Canvas Layers

## Overview

This ticket captures the design analysis for replacing the current inline autocomplete popup in `pkg/repl` with a lipgloss v2 layer/canvas overlay.

The analysis includes:

- current-state code mapping (`pkg/repl/model.go`, `config.go`, `keymap.go`, `styles.go`),
- grail-js reference architecture for layered composition,
- lipgloss v2 API/source behavior validation,
- recommended implementation plan for max-size clamping and viewport paging/scrolling.

## Key Links

- **Analysis document**: [design/01-autocomplete-overlay-with-lipgloss-v2-canvas-layers-analysis-and-design.md](./design/01-autocomplete-overlay-with-lipgloss-v2-canvas-layers-analysis-and-design.md)
- **Research diary**: [reference/01-diary.md](./reference/01-diary.md)
- **External discussion**: <https://github.com/charmbracelet/lipgloss/discussions/506>
- **reMarkable upload target**: `/ai/2026/02/13/BOBA-006-COMPLETION-OVERLAY-LIPGLOSS-V2`

## Status

Current status: **active**

## Topics

- analysis
- repl
- autocomplete
- lipgloss

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
