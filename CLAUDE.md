# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests (requires Docker for PostgreSQL containers)
make test
# or: go test -p 8 -timeout 60s ./...

# Run a single test
go test -run TestCelestialsTestSuite/TestSync ./pkg/module/...

# Lint
make lint

# Regenerate mocks and enums
make generate
```

## Architecture

This is a **Celestia blockchain indexing module** built on the [DipDup indexer SDK](https://github.com/dipdup-net/indexer-sdk). It scans for changes to Celestial IDs (blockchain domains) from an external API and persists them to PostgreSQL.

### Data flow

1. `Module` (pkg/module) runs a periodic sync loop (default: 1 min interval)
2. Fetches changes from external API (`pkg/api/v1`) starting from the last known `change_id`
3. Writes results to PostgreSQL in a transaction: upsert celestials → update state → flush
4. `CelestialState` table tracks the last processed `change_id` to enable resumable syncing

### Key packages

- **`pkg/module`** — Core orchestration. `Module` embeds DipDup's `BaseModule` and implements the sync loop. Configured via `ModuleOption` functions (`WithIndexPeriod`, `WithLimit`, `WithDatabaseTimeout`). Accepts an `AddressHandler` callback to resolve blockchain addresses.
- **`pkg/api`** — External Celestials API client interface + `v1/` HTTP implementation using fast-shot (rate-limited to 5 req/s).
- **`pkg/storage`** — Storage interfaces (`ICelestial`, `ICelestialState`, `CelestialTransaction`) and status enum (`NOT_VERIFIED`, `VERIFIED`, `PRIMARY`).
- **`pkg/storage/postgres`** — Bun ORM implementation. Uses PostgreSQL `ON CONFLICT DO UPDATE` for idempotent upserts. Status transitions: `PRIMARY → VERIFIED` when a new primary address is set for an existing celestial.

### Testing

Tests use `testcontainers` to spin up a real PostgreSQL 15.8 + TimescaleDB instance. Fixtures are loaded from `/test/*.yml` via go-testfixtures. Mocks for all interfaces are auto-generated via mockgen (stored in `pkg/*/mock/`).

When adding new interfaces, add a `//go:generate mockgen ...` directive and run `make generate`.
