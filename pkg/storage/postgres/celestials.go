package postgres

import (
	"context"

	"github.com/celenium-io/celestial-module/pkg/storage"
	"github.com/dipdup-net/go-lib/database"
)

type Celestials struct {
	*database.Bun
}

func NewCelestials(db *database.Bun) *Celestials {
	return &Celestials{
		Bun: db,
	}
}

func (c *Celestials) ById(ctx context.Context, id string) (result storage.Celestial, err error) {
	err = c.DB().NewSelect().
		Model(&result).
		Where("id = ?", id).
		Limit(1).
		Scan(ctx)
	return
}

func (c *Celestials) ByAddressId(ctx context.Context, addressId uint64, limit, offset int) (result []storage.Celestial, err error) {
	query := c.DB().NewSelect().
		Model(&result).
		Where("address_id = ?", addressId).
		Offset(offset).
		OrderExpr("change_id desc")

	if limit < 0 || limit > 100 {
		limit = 10
	}

	err = query.Limit(limit).Scan(ctx)
	return
}

func (c *Celestials) Primary(ctx context.Context, addressId uint64) (result storage.Celestial, err error) {
	err = c.DB().NewSelect().
		Model(&result).
		Where("address_id = ?", addressId).
		Where("status = ?", storage.StatusPRIMARY).
		Scan(ctx)
	return
}
