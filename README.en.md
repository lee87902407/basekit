# Basekit

## Project Purpose

`basekit` is a shared foundational library for reusable capabilities across projects, such as protocol parsing, memory pooling, caching, utilities, and infrastructure-level helpers.

## Current Status

The repository now includes its first formal module, `mempool`, and additional foundational modules will be added incrementally.

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
- optional `Buffer` wrapper objects
- request-scoped batch cleanup through `Scope`
- optional runtime misuse checks behind the `debug` build tag

Links:

- Design doc: [`docs/designs/memory-pool-design.md`](./docs/designs/memory-pool-design.md)
- Example doc: [`docs/examples/mempool.md`](./docs/examples/mempool.md)
- Example code: [`examples/mempool/`](./examples/mempool/)

Behavior notes:

- In the default build, `Buffer` does not panic on use-after-release or double-release checks; if it is used again after release, it automatically becomes managed again so a later `Scope.Close()` can still reclaim it.
- When built or tested with `-tags debug`, `Buffer` enables runtime safety checks so misuse can fail fast during development and verification.

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
- Design doc: [`docs/designs/memory-pool-design.md`](./docs/designs/memory-pool-design.md)
- Example documents: [`docs/examples/README.md`](./docs/examples/README.md)
- Example code guide: [`examples/README.md`](./examples/README.md)
- AI collaboration rules: [`AGENTS.md`](./AGENTS.md)
