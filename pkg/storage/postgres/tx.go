package postgres

import (
	"context"
	"iter"

	"github.com/celenium-io/celestial-module/pkg/storage"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

type CelestialTransaction struct {
	sdk.Transaction
}

func BeginCelestialTransaction(ctx context.Context, tx sdk.Transactable) (CelestialTransaction, error) {
	t, err := tx.BeginTransaction(ctx)
	return CelestialTransaction{t}, err
}

func (tx CelestialTransaction) SaveCelestials(ctx context.Context, celestials iter.Seq[storage.Celestial]) error {
	for cel := range celestials {
		_, err := tx.Tx().NewInsert().
			Model(&cel).
			Column("id", "address_id", "image_url", "change_id").
			On("CONFLICT (id) DO UPDATE").
			Set("address_id = EXCLUDED.address_id").
			Set("image_url = EXCLUDED.image_url").
			Set("change_id = EXCLUDED.change_id").
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tx CelestialTransaction) UpdateState(ctx context.Context, state *storage.CelestialState) error {
	_, err := tx.Tx().NewUpdate().
		Model(state).
		Set("change_id = ?", state.ChangeId).
		WherePK().
		Exec(ctx)
	return err
}
