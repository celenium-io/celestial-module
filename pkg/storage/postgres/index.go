package postgres

import (
	"context"

	"github.com/celenium-io/celestial-module/pkg/storage"
	"github.com/uptrace/bun"
)

// CreateIndex - creates all needed indices in postgres database
func CreateIndex(ctx context.Context, tx bun.Tx) error {
	if _, err := tx.NewCreateIndex().
		IfNotExists().
		Model((*storage.Celestial)(nil)).
		Index("celestial_address_id_idx").
		Column("address_id").
		Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.NewCreateIndex().
		IfNotExists().
		Model((*storage.Celestial)(nil)).
		Index("celestial_change_id_idx").
		Column("change_id").
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
