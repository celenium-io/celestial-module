# Celestial Module

A Go module for indexing [Celestial IDs](https://celestials.id/) — domain names on the Celestia network. Built on top of the [DipDup Indexer SDK](https://github.com/dipdup-net/indexer-sdk) and designed to be embedded into an indexer as a standalone module.

## What it does

The module periodically polls the external Celestials API, fetches domain changes (registrations, address updates, status changes), and persists them to PostgreSQL. It tracks progress using a `change_id` — the module remembers the last processed identifier and resumes from that point on restart.

## Requirements

- Go 1.25+
- PostgreSQL 15+ with the TimescaleDB extension
- Docker (for running tests)

## Installation

```bash
go get github.com/celenium-io/celestial-module
```

## Usage

```go
import (
    "github.com/celenium-io/celestial-module/pkg/module"
    "github.com/celenium-io/celestial-module/pkg/storage"
    "github.com/celenium-io/celestial-module/pkg/storage/postgres"
)

m := module.New(
    celestialsDatasource, // config.DataSource with Celestials API URL
    addressHandler,       // func(ctx, address string) (uint64, error)
    celestialsStorage,    // storage.ICelestial
    stateStorage,         // storage.ICelestialState
    transactable,         // sdk.Transactable
    "my-indexer",         // indexer name for state tracking
    "celestia",           // network name
    module.WithIndexPeriod(30*time.Second),    // sync interval (default: 1 min)
    module.WithLimit(200),                     // batch size (default: 100)
    module.WithDatabaseTimeout(2*time.Minute), // DB operation timeout (default: 1 min)
)

m.Start(ctx)
defer m.Close()
```

`AddressHandler` is a callback the module uses to resolve a string address into an internal address ID. It should be implemented on the indexer side.

## Structure

```
pkg/
├── api/            # External Celestials API client
│   ├── v1/         # HTTP implementation (fast-shot, rate limited to 5 req/s)
│   └── mock/       # Auto-generated mocks
├── module/         # Core indexing module
└── storage/        # Storage interfaces and data models
    ├── postgres/   # Bun ORM implementation (PostgreSQL)
    └── mock/       # Auto-generated mocks
```

## Data models

**Celestial** — a domain name:

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Domain identifier (PK) |
| `address_id` | uint64 | Internal ID of the linked address |
| `image_url` | string | Image URL |
| `change_id` | int64 | ID of the last change |
| `status` | enum | `NOT_VERIFIED`, `VERIFIED`, `PRIMARY` |

**CelestialState** — sync state:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Indexer name (PK) |
| `change_id` | int64 | Last processed change ID |

## Development

```bash
# Run tests (requires Docker)
make test

# Lint
make lint

# Regenerate mocks and enum methods
make generate
```

## License

[MIT](LICENSE)
