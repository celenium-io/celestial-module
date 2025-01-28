package storage

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ICelestialState interface {
	ByName(ctx context.Context, name string) (CelestialState, error)
	Save(ctx context.Context, state *CelestialState) error
}

type CelestialState struct {
	bun.BaseModel `bun:"celestial_state" comment:"Table with celestial ids."`

	Name     string `bun:"name,pk,notnull" comment:"Celestial id indexer name"`
	ChangeId int64  `bun:"change_id"       comment:"Id of the last change of celestial id"`
}

func (CelestialState) TableName() string {
	return "celestial_state"
}

func (cid CelestialState) String() string {
	return fmt.Sprintf("%s %d", cid.Name, cid.ChangeId)
}
