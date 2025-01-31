package postgres

import (
	"context"
	"database/sql"
	"slices"
	"testing"
	"time"

	"github.com/celenium-io/celestial-module/pkg/storage"
	"github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
)

// CelestialsTestSuite -
type CelestialsTestSuite struct {
	suite.Suite
	psqlContainer *database.PostgreSQLContainer
	storage       *postgres.Storage

	celestials     *Celestials
	celestialState *CelestialState
}

// SetupSuite -
func (s *CelestialsTestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer ctxCancel()

	psqlContainer, err := database.NewPostgreSQLContainer(ctx, database.PostgreSQLContainerConfig{
		User:     "user",
		Password: "password",
		Database: "db_test",
		Port:     5432,
		Image:    "timescale/timescaledb-ha:pg15.8-ts2.17.0-all",
	})
	s.Require().NoError(err)
	s.psqlContainer = psqlContainer

	init := func(ctx context.Context, conn *database.Bun) error {
		if err := CreateTypes(ctx, conn); err != nil {
			return err
		}
		if err := database.CreateTables(ctx, conn, new(storage.Celestial), new(storage.CelestialState)); err != nil {
			if err := conn.Close(); err != nil {
				return err
			}
			return err
		}
		return nil
	}

	strg, err := postgres.Create(ctx, config.Database{
		Kind:     config.DBKindPostgres,
		User:     s.psqlContainer.Config.User,
		Database: s.psqlContainer.Config.Database,
		Password: s.psqlContainer.Config.Password,
		Host:     s.psqlContainer.Config.Host,
		Port:     s.psqlContainer.MappedPort().Int(),
	}, init)
	s.Require().NoError(err)
	s.storage = strg

	s.celestials = NewCelestials(strg.Connection())
	s.celestialState = NewCelestialState(strg.Connection())

	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("timescaledb"),
		testfixtures.Directory("../../../test"),
		testfixtures.UseAlterConstraint(),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
	s.Require().NoError(db.Close())
}

// TearDownSuite -
func (s *CelestialsTestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.storage.Close())
	s.Require().NoError(s.psqlContainer.Terminate(ctx))
}

func TestSuiteCelestials_Run(t *testing.T) {
	suite.Run(t, new(CelestialsTestSuite))
}

func (s *CelestialsTestSuite) TestCelestialsById() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	item, err := s.celestials.ById(ctx, "name 3")
	s.Require().NoError(err)
	s.Require().EqualValues("", item.ImageUrl)
	s.Require().EqualValues("name 3", item.Id)
	s.Require().EqualValues(3, item.ChangeId)
	s.Require().EqualValues(2, item.AddressId)
}

func (s *CelestialsTestSuite) TestCelestialsByAddressId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	items, err := s.celestials.ByAddressId(ctx, 1, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(items, 2)

	item := items[0]
	s.Require().EqualValues("", item.ImageUrl)
	s.Require().EqualValues("name 2", item.Id)
	s.Require().EqualValues(2, item.ChangeId)
	s.Require().EqualValues(1, item.AddressId)
}

func (s *CelestialsTestSuite) TestCelestialsPrimary() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	item, err := s.celestials.Primary(ctx, 1)
	s.Require().NoError(err)

	s.Require().EqualValues("", item.ImageUrl)
	s.Require().EqualValues("name 1", item.Id)
	s.Require().EqualValues(1, item.ChangeId)
	s.Require().EqualValues(1, item.AddressId)
}

func (s *CelestialsTestSuite) TestCelestialsTransaction() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginCelestialTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	state, err := s.celestialState.ByName(ctx, "indexer")
	s.Require().NoError(err)

	celIds := []storage.Celestial{
		{
			Id:        "name 3",
			AddressId: 1,
			ChangeId:  4,
			ImageUrl:  "image_url",
			Status:    storage.StatusPRIMARY,
		}, {
			Id:        "name 4",
			AddressId: 3,
			ChangeId:  5,
			ImageUrl:  "image_url2",
			Status:    storage.StatusPRIMARY,
		},
	}
	state.ChangeId = celIds[1].ChangeId

	err = tx.UpdateStatusForAddress(ctx, slices.Values([]uint64{1, 3}))
	s.Require().NoError(err)

	err = tx.SaveCelestials(ctx, slices.Values(celIds))
	s.Require().NoError(err)

	err = tx.UpdateState(ctx, &state)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	state1, err := s.celestialState.ByName(ctx, "indexer")
	s.Require().NoError(err)
	s.Require().EqualValues(celIds[1].ChangeId, state1.ChangeId)

	item, err := s.celestials.ById(ctx, "name 3")
	s.Require().NoError(err)
	s.Require().EqualValues("image_url", item.ImageUrl)
	s.Require().EqualValues("name 3", item.Id)
	s.Require().EqualValues(4, item.ChangeId)
	s.Require().EqualValues(1, item.AddressId)
	s.Require().EqualValues(storage.StatusPRIMARY, item.Status)

	item2, err := s.celestials.ById(ctx, "name 4")
	s.Require().NoError(err)
	s.Require().EqualValues("image_url2", item2.ImageUrl)
	s.Require().EqualValues("name 4", item2.Id)
	s.Require().EqualValues(5, item2.ChangeId)
	s.Require().EqualValues(3, item2.AddressId)
	s.Require().EqualValues(storage.StatusPRIMARY, item.Status)

	item3, err := s.celestials.ById(ctx, "name 1")
	s.Require().NoError(err)
	s.Require().EqualValues("", item3.ImageUrl)
	s.Require().EqualValues("name 1", item3.Id)
	s.Require().EqualValues(1, item3.ChangeId)
	s.Require().EqualValues(1, item3.AddressId)
	s.Require().EqualValues(storage.StatusVERIFIED, item3.Status)
}

func (s *CelestialsTestSuite) TestCelestialsStateSave() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	err := s.celestialState.Save(ctx, &storage.CelestialState{
		Name:     "new",
		ChangeId: 10,
	})
	s.Require().NoError(err)

	state, err := s.celestialState.ByName(ctx, "new")
	s.Require().NoError(err)
	s.Require().EqualValues(10, state.ChangeId)
	s.Require().EqualValues("new", state.Name)
}

func (s *CelestialsTestSuite) TestCelestialsStateByName() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	state, err := s.celestialState.ByName(ctx, "indexer")
	s.Require().NoError(err)
	s.Require().EqualValues(3, state.ChangeId)
	s.Require().EqualValues("indexer", state.Name)
}
