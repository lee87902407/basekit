# Basekit

## Project Purpose

`basekit` is a shared foundational library for reusable capabilities across projects, such as protocol parsing, memory pooling, caching, utilities, and infrastructure-level helpers.

## Current Status

The repository is initialized with documentation and collaboration conventions. Functional modules will be added incrementally.

## Structure

- `README.md`: unified entrypoint for navigation.
- `README.zh-CN.md`: Chinese primary documentation.
- `README.en.md`: English guide.
- `AGENTS.md`: AI collaboration and repository maintenance rules.
- `docs/examples/`: per-module example documents.
- `examples/`: per-module example code.

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
- Example documents: [`docs/examples/README.md`](./docs/examples/README.md)
- Example code guide: [`examples/README.md`](./examples/README.md)
- AI collaboration rules: [`AGENTS.md`](./AGENTS.md)
