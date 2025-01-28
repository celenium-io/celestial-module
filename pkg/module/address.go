package module

import "context"

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IdByHash interface {
	IdByHash(ctx context.Context, hash ...[]byte) ([]uint64, error)
}
