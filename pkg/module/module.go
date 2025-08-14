package module

import (
	"context"
	"database/sql"
	"maps"
	"time"

	celestials "github.com/celenium-io/celestial-module/pkg/api"
	v1 "github.com/celenium-io/celestial-module/pkg/api/v1"
	"github.com/celenium-io/celestial-module/pkg/storage"
	"github.com/celenium-io/celestial-module/pkg/storage/postgres"
	"github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/indexer-sdk/pkg/modules"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type AddressHandler func(ctx context.Context, address string) (uint64, error)

type Module struct {
	modules.BaseModule

	celestialsApi  celestials.API
	addressHandler AddressHandler
	states         storage.ICelestialState
	celestials     storage.ICelestial
	tx             sdk.Transactable
	state          storage.CelestialState

	celestialsDatasource config.DataSource
	indexerName          string
	network              string
	indexPeriod          time.Duration
	databaseTimeout      time.Duration
	limit                int64
}

func New(
	celestialsDatasource config.DataSource,
	addressHandler AddressHandler,
	celestials storage.ICelestial,
	state storage.ICelestialState,
	tx sdk.Transactable,
	indexerName string,
	network string,
	opts ...ModuleOption,
) *Module {
	module := Module{
		BaseModule:           modules.New("celestials"),
		celestials:           celestials,
		states:               state,
		tx:                   tx,
		celestialsApi:        v1.New(celestialsDatasource.URL),
		indexerName:          indexerName,
		network:              network,
		indexPeriod:          time.Minute,
		databaseTimeout:      time.Minute,
		limit:                100,
		celestialsDatasource: celestialsDatasource,
		addressHandler:       addressHandler,
	}

	for i := range opts {
		opts[i](&module)
	}

	return &module
}

func (m *Module) Close() error {
	m.Log.Info().Msg("closing scanner...")
	m.G.Wait()

	return nil
}

func (m *Module) Start(ctx context.Context) {
	if m.addressHandler == nil {
		panic("nil address handler")
	}
	if err := m.getState(ctx); err != nil {
		m.Log.Err(err).Msg("state receiving")
		return
	}
	m.Log.Info().Msg("starting scanner...")
	m.G.GoCtx(ctx, m.receive)
}

func (m *Module) getState(ctx context.Context) error {
	requestCtx, cancel := context.WithTimeout(ctx, m.databaseTimeout)
	defer cancel()

	state, err := m.states.ByName(requestCtx, m.indexerName)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return errors.Wrap(err, "state by name")
		}
		m.state = storage.CelestialState{
			Name:     m.indexerName,
			ChangeId: 0,
		}
		return m.states.Save(ctx, &m.state)
	}
	m.state = state
	return nil
}

func (m *Module) receive(ctx context.Context) {
	if err := m.sync(ctx); err != nil {
		m.Log.Err(err).Msg("sync")
	}

	ticker := time.NewTicker(m.indexPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.sync(ctx); err != nil {
				m.Log.Err(err).Msg("sync")
			}
		}
	}
}

func (m *Module) getChanges(ctx context.Context) (celestials.Changes, error) {
	requestCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(m.celestialsDatasource.Timeout))
	defer cancel()

	return m.celestialsApi.Changes(
		requestCtx,
		m.network,
		celestials.WithFromChangeId(m.state.ChangeId),
		celestials.WithImages(),
		celestials.WithLimit(m.limit),
	)
}

func (m *Module) sync(ctx context.Context) error {
	m.Log.Debug().Msg("start syncing...")

	var end bool

	for !end {
		changes, err := m.getChanges(ctx)
		if err != nil {
			return errors.Wrap(err, "get changes")
		}
		log.Info().
			Int("changes_count", len(changes.Changes)).
			Int64("head", changes.Head).
			Msg("received changes")

		cids := make(map[string]storage.Celestial)
		addressIds := make(map[uint64]struct{})

		var lastId int64
		for i := range changes.Changes {
			if m.state.ChangeId >= changes.Changes[i].ChangeID {
				continue
			}
			lastId = changes.Changes[i].ChangeID

			status, err := storage.ParseStatus(changes.Changes[i].Status)
			if err != nil {
				return err
			}
			addressId, err := m.addressHandler(ctx, changes.Changes[i].Address)
			if err != nil {
				m.Log.Err(err).Msg("address handler")
				continue
			}

			if status == storage.StatusPRIMARY {
				addressIds[addressId] = struct{}{}
			}

			cids[changes.Changes[i].CelestialID] = storage.Celestial{
				Id:        changes.Changes[i].CelestialID,
				ImageUrl:  changes.Changes[i].ImageURL,
				AddressId: addressId,
				ChangeId:  changes.Changes[i].ChangeID,
				Status:    status,
			}
		}

		if lastId > m.state.ChangeId {
			m.state.ChangeId = lastId

			if err := m.save(ctx, cids, addressIds); err != nil {
				return errors.Wrap(err, "save")
			}
			log.Debug().
				Int("changes_count", len(cids)).
				Int64("head", m.state.ChangeId).
				Msg("saved changes")
		}

		end = len(changes.Changes) < int(m.limit)
	}

	m.Log.Debug().Msg("end syncing...")
	return nil
}

func (m *Module) save(ctx context.Context, cids map[string]storage.Celestial, addressIds map[uint64]struct{}) error {
	requestCtx, cancel := context.WithTimeout(ctx, m.databaseTimeout)
	defer cancel()

	tx, err := postgres.BeginCelestialTransaction(requestCtx, m.tx)
	if err != nil {
		return errors.Wrap(err, "begin transactions")
	}
	defer tx.Close(requestCtx)

	if err := tx.UpdateStatusForAddress(ctx, maps.Keys(addressIds)); err != nil {
		return tx.HandleError(requestCtx, errors.Wrap(err, "update primary statuses"))
	}

	if err := tx.SaveCelestials(requestCtx, maps.Values(cids)); err != nil {
		return tx.HandleError(requestCtx, errors.Wrap(err, "save celestials"))
	}

	if err := tx.UpdateState(requestCtx, &m.state); err != nil {
		return tx.HandleError(requestCtx, errors.Wrap(err, "update state"))
	}

	return tx.Flush(requestCtx)
}
