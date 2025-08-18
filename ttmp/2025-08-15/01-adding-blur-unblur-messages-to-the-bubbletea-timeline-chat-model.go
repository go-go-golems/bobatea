package plan

/*
Plan: Adding BlurInput/UnblurInput messages to the Bubble Tea timeline chat model

Purpose and scope
- Enable UI entities to explicitly control the input field behavior in the Bubble Tea chat UI implemented in `bobatea/pkg/chat/model.go`.
- Introduce two new messages: `BlurInputMsg` and `UnblurInputMsg` that can be dispatched by UI entities through the existing `UserActionMsg` pathway.
- Wire these messages into the model so that the input can be programmatically blurred or unblurred without requiring direct user interaction.
- Optionally gate input processing when blurred and provide a clean API for UI controllers.

Context
- The chat model currently uses `m.textArea` for input, with calls to `Focus()` and `Blur()` in response to various actions. It routes user actions through `handleUserAction` via the `UserActionMsg` interface. See `bobatea/pkg/chat/model.go` for current patterns.

What to change (high-level)
- Add two new messages implementing `UserActionMsg`:
  - `BlurInputMsg struct{}`
  - `UnblurInputMsg struct{}`
- Extend `model` with an optional blur flag (e.g., `inputBlurred bool`) to enable gating if desired.
- Extend `handleUserAction` to handle these two messages, invoking `m.textArea.Blur()` and `m.textArea.Focus()` respectively and updating the blur flag.
- Wire message dispatch through the existing `Update` / `handleUserAction` path by ensuring `BlurInputMsg` and `UnblurInputMsg` satisfy the `UserActionMsg` interface (so they land in the same switch and are correctly routed).
- (Optional) Gate keyboard input when blurred by adding a guard in the key handling path.
- Document a minimal UI integration path so UI entities can trigger these messages (via the timeline or a dedicated UI action path).

Implementation plan (step-by-step)
- [ ] 1) Define the messages
  - [ ] Create `BlurInputMsg` and `UnblurInputMsg` types that implement `UserActionMsg`.
  - [ ] Add small, self-contained comments describing their purpose.
  - Pseudo-code:
  ```go
  // BlurInputMsg is dispatched by UI to blur the input field
  type BlurInputMsg struct{}

  // UnblurInputMsg is dispatched by UI to re-enable/focus the input field
  type UnblurInputMsg struct{}
  ```
- [ ] 2) Extend the model to track blur state (optional)
  - [ ] Add `inputBlurred bool` field to `type model struct { ... }` and default to false.
  - Pseudo-code:
  ```go
  type model struct {
    // ... existing fields
    inputBlurred bool
  }
  ```
- [ ] 3) Wire message handling in handleUserAction
  - [ ] In the switch over concrete types, add cases for `BlurInputMsg` and `UnblurInputMsg`.
  - Pseudo-code:
  ```go
  switch msg_ := msg.(type) {
  // existing cases...
  case BlurInputMsg:
      m.textArea.Blur()
      m.inputBlurred = true
      return m, nil
  case UnblurInputMsg:
      m.textArea.Focus()
      m.inputBlurred = false
      return m, nil
  }
  ```
- [ ] 4) Ensure Update routes through UserActionMsg path
  - [ ] Ensure `BlurInputMsg` and `UnblurInputMsg` are treated as `UserActionMsg` so they reach `handleUserAction` via the existing `case UserActionMsg` branch.
- [ ] 5) Optional: Gate input processing when blurred
  - [ ] If `inputBlurred` is true, skip forwarding KeyMsg to `m.textArea.Update` (or adjust as desired by UX).
  - Pseudo-code:
  ```go
  if m.inputBlurred && msg_ is KeyMsg {
      // ignore input when blurred
      return m, nil
  }
  ```
- [ ] 6) UI integration notes
  - [ ] Propose a simple dispatch pattern for UI controllers that want to blur/unblur via the existing action mechanism:
  - [ ] UI action handler -> return `BlurInputMsg{}` or `UnblurInputMsg{}` as a `UserActionMsg` to be processed by the model.
- [ ] 7) Testing and validation
  - [ ] Build the bobatea module and ensure no compilation errors.
  - [ ] Add a small unit-test idea (optional in repo) to simulate dispatching `BlurInputMsg`/`UnblurInputMsg` and asserting `m.textArea` state transitions.
- [ ] 8) Documentation / traceability
  - [ ] Document the new path in ttmp, and consider a short guide for UI developers to trigger these messages.

Risks and considerations
- Minimal breaking surface as this adds new message types and optional gating; existing flows should remain unaffected.
- If the UI path to dispatch these messages relies on a broader timeline plumbing, adapt the plan to the project’s actual messaging surface (likely via `UserActionMsg`).

Success criteria
- The new messages exist and are routed to the chat model.
- The input field can be programmatically blurred and unblurred via messages.
- The UI can trigger these messages without changing the current user experience for non-blurred input.

*/
