package postgres

import (
	"context"
	"iter"
	"slices"

	"github.com/celenium-io/celestial-module/pkg/storage"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
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
			Column("id", "address_id", "image_url", "change_id", "status").
			On("CONFLICT (id) DO UPDATE").
			Set("address_id = EXCLUDED.address_id").
			Set("image_url = EXCLUDED.image_url").
			Set("change_id = EXCLUDED.change_id").
			Set("status = EXCLUDED.status").
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

func (tx CelestialTransaction) UpdateStatusForAddress(ctx context.Context, addressId iter.Seq[uint64]) error {
	_, err := tx.Tx().NewUpdate().
		Model((*storage.Celestial)(nil)).
		Set("status = ?", storage.StatusVERIFIED).
		Where("address_id IN (?)", bun.In(slices.Collect(addressId))).
		Where("status = ?", storage.StatusPRIMARY).
		Exec(ctx)
	return err
}
