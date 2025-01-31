package module

import "time"

type ModuleOption func(*Module)

func WithIndexPeriod(period time.Duration) ModuleOption {
	return func(m *Module) {
		m.indexPeriod = period
	}
}

func WithDatabaseTimeout(timeout time.Duration) ModuleOption {
	return func(m *Module) {
		m.databaseTimeout = timeout
	}
}

func WithLimit(limit int64) ModuleOption {
	return func(m *Module) {
		m.limit = limit
	}
}
