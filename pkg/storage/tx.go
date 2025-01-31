package storage

import (
	"context"
	"iter"

	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type CelestialTransaction interface {
	SaveCelestials(ctx context.Context, celestials iter.Seq[Celestial]) error
	UpdateState(ctx context.Context, state *CelestialState) error
	UpdateStatusForAddress(ctx context.Context, addressId ...iter.Seq[uint64]) error

	sdk.Transaction
}
