package postgres

import (
	"context"

	"github.com/celenium-io/celestial-module/pkg/storage"
	"github.com/dipdup-net/go-lib/database"
)

type CelestialState struct {
	db *database.Bun
}

func NewCelestialState(db *database.Bun) *CelestialState {
	return &CelestialState{
		db: db,
	}
}

func (cs *CelestialState) ByName(ctx context.Context, name string) (result storage.CelestialState, err error) {
	err = cs.db.DB().NewSelect().
		Model(&result).
		Where("name = ?", name).
		Limit(1).
		Scan(ctx)
	return
}

func (cs *CelestialState) Save(ctx context.Context, state *storage.CelestialState) error {
	_, err := cs.db.DB().NewInsert().
		Model(state).
		Exec(ctx)
	return err
}
