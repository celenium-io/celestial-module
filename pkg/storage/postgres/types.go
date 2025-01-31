package postgres

import (
	"context"
	"database/sql"

	"github.com/celenium-io/celestial-module/pkg/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

const (
	createTypeQuery = `DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = ?) THEN
			CREATE TYPE ? AS ENUM (?);
		END IF;
	END$$;`
)

func CreateTypes(ctx context.Context, conn *database.Bun) error {
	log.Info().Msg("creating celestial types...")
	return conn.DB().RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.ExecContext(
			ctx,
			createTypeQuery,
			"celestials_status",
			bun.Safe("celestials_status"),
			bun.In(storage.StatusValues()),
		); err != nil {
			return err
		}

		return nil
	})
}
