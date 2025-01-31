package module

import (
	"context"
	"database/sql"
	"testing"
	"time"

	celestials "github.com/celenium-io/celestial-module/pkg/api"
	celestialsMock "github.com/celenium-io/celestial-module/pkg/api/mock"
	"github.com/celenium-io/celestial-module/pkg/storage"
	pg "github.com/celenium-io/celestial-module/pkg/storage/postgres"
	"github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

const testIndexerName = "indexer"
const network = "celestia"

// ModuleTestSuite -
type ModuleTestSuite struct {
	suite.Suite
	psqlContainer *database.PostgreSQLContainer
	storage       *postgres.Storage

	celestials     *pg.Celestials
	celestialState *pg.CelestialState
	ctrl           *gomock.Controller
	api            *celestialsMock.MockAPI
}

// SetupSuite -
func (s *ModuleTestSuite) SetupSuite() {
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
		if err := pg.CreateTypes(ctx, conn); err != nil {
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

	s.celestials = pg.NewCelestials(strg.Connection())
	s.celestialState = pg.NewCelestialState(strg.Connection())

	s.ctrl = gomock.NewController(s.T())
	s.api = celestialsMock.NewMockAPI(s.ctrl)
}

// TearDownSuite -
func (s *ModuleTestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.storage.Close())
	s.Require().NoError(s.psqlContainer.Terminate(ctx))
	s.ctrl.Finish()
}

func (s *ModuleTestSuite) TestSync() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("timescaledb"),
		testfixtures.Directory("../../test"),
		testfixtures.UseAlterConstraint(),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
	s.Require().NoError(db.Close())

	s.api.EXPECT().
		Changes(
			gomock.Any(),
			network,
			gomock.Any(),
		).
		Times(1).
		Return(celestials.Changes{
			Head: 3,
			Changes: []celestials.Change{
				{
					CelestialID: "test",
					Address:     "celestia1mm8yykm46ec3t0dgwls70g0jvtm055wk9ayal8",
					ImageURL:    "image_url",
					ChangeID:    4,
					Status:      "PRIMARY",
				},
			},
		}, nil)

	cfgDs := config.DataSource{
		Kind:              "celestials",
		URL:               "base_url",
		Timeout:           10,
		RequestsPerSecond: 10,
	}

	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ctxCancel()

	m := New(
		cfgDs,
		func(ctx context.Context, address string) (uint64, error) {
			return 1, nil
		},
		s.celestials,
		s.celestialState,
		s.storage.Transactable,
		testIndexerName,
		network,
		WithLimit(10),
	)
	m.celestialsApi = s.api

	err = m.getState(ctx)
	s.Require().NoError(err)

	err = m.sync(ctx)
	s.Require().NoError(err)

	st, err := s.celestialState.ByName(ctx, testIndexerName)
	s.Require().NoError(err)
	s.Require().EqualValues(4, st.ChangeId)
	s.Require().EqualValues(testIndexerName, st.Name)

	item, err := s.celestials.ById(ctx, "test")
	s.Require().NoError(err)
	s.Require().EqualValues("image_url", item.ImageUrl)
	s.Require().EqualValues("test", item.Id)
	s.Require().EqualValues(4, item.ChangeId)
	s.Require().EqualValues(1, item.AddressId)
}

func TestSuiteModule_Run(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}
