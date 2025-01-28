package storage

import (
	"context"

	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type CelestialTransaction interface {
	SaveCelestials(ctx context.Context, celestials ...Celestial) error
	UpdateState(ctx context.Context, state *CelestialState) error

	sdk.Transaction
}
