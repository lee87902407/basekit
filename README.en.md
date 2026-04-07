# Basekit

## Project Purpose

`basekit` is a shared foundational library for reusable capabilities across projects, such as protocol parsing, memory pooling, caching, utilities, and infrastructure-level helpers.

## Current Status

The repository now includes formal modules `mempool` and `log`, and additional foundational modules will be added incrementally.

## Structure

- `README.md`: unified entrypoint for navigation.
- `README.zh-CN.md`: Chinese primary documentation.
- `README.en.md`: English guide.
- `AGENTS.md`: AI collaboration and repository maintenance rules.
- `docs/designs/`: formal design documents.
- `docs/examples/`: per-module example documents.
- `examples/`: per-module example code.

## Current Module

### mempool

`mempool` is a bucketed `[]byte` memory pool built on top of `sync.Pool`, designed for high-frequency short-lived buffer workloads. It supports:

- bucketed reuse up to 512KB
- exact allocation and drop-on-put for oversized buffers
- a shared `Buffer` interface
- `Buffer.Type()` plus `BufferTypeHeap` / `BufferTypeNative` constants
- writable `HeapBuffer` wrapper objects
- read-only `NativeBuffer` wrappers for cgo-backed memory
- request-scoped batch cleanup through `Scope`
- optional runtime misuse checks behind the `debug` build tag

Links:

- Design doc: [`docs/designs/memory-pool-design.md`](./docs/designs/memory-pool-design.md)
- Example doc: [`docs/examples/mempool.md`](./docs/examples/mempool.md)
- Example code: [`examples/mempool/`](./examples/mempool/)

Behavior notes:

- `mempool.NewHeapBuffer` creates the pooled writable implementation, while `Scope.NewBuffer` continues to return `*HeapBuffer`, which implements the shared `Buffer` interface.
- When cgo is enabled, `mempool.NewNativeBuffer`, `Scope.NewNativeBuffer`, and `Scope.GetNativeBuffer` wrap native memory as a read-only `Buffer`; every write-style method panics immediately, and release is delegated to the injected `freeFun` callback.
- `Scope.GetHeapBuffer` allocates and tracks raw heap `[]byte`; the old `Track` entrypoint has been removed.
- In the default build, `HeapBuffer` does not panic on use-after-release or double-release checks; if it is used again after release, it automatically becomes managed again so a later `Scope.Close()` can still reclaim it.
- When built or tested with `-tags debug`, `Buffer` lifecycle tracking enables runtime safety checks so misuse can fail fast during development and verification.
- Once `Scope.Close()` has been called, subsequent calls to `GetHeapBuffer`, `NewBuffer`, `NewNativeBuffer`, or `GetNativeBuffer` panic to prevent resources from escaping the scope cleanup path.

### log

`log` is a global singleton logging module built on top of `zap`, intended for service startup logs, structured business logs, and local debugging workflows. It supports:

- global `Init(Config)` initialization
- package-level `Debug` / `Info` / `Warn` / `Error` / `Sync`
- `OutputModeConsole` / `OutputModeFile` / `OutputModeBoth`
- console text output and file JSON output
- size-based file rotation and retention through `lumberjack`
- runtime level changes through `SetLevel`
- fallback console logging before explicit initialization

Links:

- Design doc: [`docs/designs/log-design.md`](./docs/designs/log-design.md)
- Example doc: [`docs/examples/log.md`](./docs/examples/log.md)
- Example code: [`examples/log/`](./examples/log/)

Behavior notes:

- `Init` is allowed to succeed only once; repeated calls return an error.
- If business code calls `Debug` / `Info` / `Warn` / `Error` before `Init`, the module falls back to a default console logger instead of dropping the logs.
- Console output uses a text encoder, file output uses a JSON encoder, and both are active together in dual-output mode.

## Maintenance Rules

Whenever a new main module is added, the following updates are required:

1. Update `README.md` with the module summary and links.
2. Keep `README.zh-CN.md` and `README.en.md` in sync.
3. Add a module example document at `docs/examples/<module>.md`.
4. Add or update module example code under `examples/<module>/`.
5. If module APIs or behavior change, update docs and examples in the same change.

## Notes

- Comments and explanatory documentation in this repository are written in Chinese by default.
- English content is maintained for external readers and cross-team collaboration.

## Links

- Unified entry: [`README.md`](./README.md)
- Log design doc: [`docs/designs/log-design.md`](./docs/designs/log-design.md)
- Design doc: [`docs/designs/memory-pool-design.md`](./docs/designs/memory-pool-design.md)
- Example documents: [`docs/examples/README.md`](./docs/examples/README.md)
- Example code guide: [`examples/README.md`](./examples/README.md)
- AI collaboration rules: [`AGENTS.md`](./AGENTS.md)
