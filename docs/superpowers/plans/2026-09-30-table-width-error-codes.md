# Table Width and Error Codes Implementation Plan

> **For agentic workers:** Implement the checked steps in order. Keep the existing API response fields and HTTP semantics.

**Goal:** Enable draggable table column widths everywhere and make API error codes stable and useful.

**Architecture:** Shared React wrappers add resize behavior to Ant Design Table and ProTable. A central Go error mapping supplies business codes to the existing response serializer.

**Tech Stack:** React 19, TypeScript, Ant Design 6, ProComponents, Go, Echo.

---

### Task 1: Error codes

- [x] Add failing response tests for success, general HTTP errors, and wrapped domain errors.
- [x] Add a central error code map and update `echox.Response.JSON`.
- [x] Run focused and full Go tests. Full suite has unrelated failures in `pkg/file` and queue metrics.

### Task 2: Resizable tables

- [x] Add a shared header resize component and wrappers for Table and ProTable.
- [x] Switch every page table to the shared wrappers.
- [x] Run TypeScript and build checks; verify all table sites are covered. Existing TypeScript errors remain outside changed files.

### Task 3: Documentation and review

- [x] Document code meanings and resizing behavior.
- [x] Review changes and report remaining verification limits.
