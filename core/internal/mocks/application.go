// Code generated by mockery v2.8.0. DO NOT EDIT.

package mocks

import (
	context "context"

	config "github.com/smartcontractkit/chainlink/core/store/config"

	eth "github.com/smartcontractkit/chainlink/core/services/eth"

	feeds "github.com/smartcontractkit/chainlink/core/services/feeds"

	health "github.com/smartcontractkit/chainlink/core/services/health"

	job "github.com/smartcontractkit/chainlink/core/services/job"

	keystore "github.com/smartcontractkit/chainlink/core/services/keystore"

	logger "github.com/smartcontractkit/chainlink/core/logger"

	mock "github.com/stretchr/testify/mock"

	null "gopkg.in/guregu/null.v4"

	packr "github.com/gobuffalo/packr"

	pipeline "github.com/smartcontractkit/chainlink/core/services/pipeline"

	store "github.com/smartcontractkit/chainlink/core/store"

	types "github.com/smartcontractkit/chainlink/core/services/headtracker/types"

	uuid "github.com/satori/go.uuid"

	webhook "github.com/smartcontractkit/chainlink/core/services/webhook"

	zapcore "go.uber.org/zap/zapcore"
)

// Application is an autogenerated mock type for the Application type
type Application struct {
	mock.Mock
}

// AddJobV2 provides a mock function with given fields: ctx, _a1, name
func (_m *Application) AddJobV2(ctx context.Context, _a1 job.Job, name null.String) (job.Job, error) {
	ret := _m.Called(ctx, _a1, name)

	var r0 job.Job
	if rf, ok := ret.Get(0).(func(context.Context, job.Job, null.String) job.Job); ok {
		r0 = rf(ctx, _a1, name)
	} else {
		r0 = ret.Get(0).(job.Job)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, job.Job, null.String) error); ok {
		r1 = rf(ctx, _a1, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteJob provides a mock function with given fields: ctx, jobID
func (_m *Application) DeleteJob(ctx context.Context, jobID int32) error {
	ret := _m.Called(ctx, jobID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int32) error); ok {
		r0 = rf(ctx, jobID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetConfig provides a mock function with given fields:
func (_m *Application) GetConfig() config.GeneralConfig {
	ret := _m.Called()

	var r0 config.GeneralConfig
	if rf, ok := ret.Get(0).(func() config.GeneralConfig); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(config.GeneralConfig)
		}
	}

	return r0
}

// GetEVMConfig provides a mock function with given fields:
func (_m *Application) GetEVMConfig() config.EVMConfig {
	ret := _m.Called()

	var r0 config.EVMConfig
	if rf, ok := ret.Get(0).(func() config.EVMConfig); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(config.EVMConfig)
		}
	}

	return r0
}

// GetEthClient provides a mock function with given fields:
func (_m *Application) GetEthClient() eth.Client {
	ret := _m.Called()

	var r0 eth.Client
	if rf, ok := ret.Get(0).(func() eth.Client); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(eth.Client)
		}
	}

	return r0
}

// GetExternalInitiatorManager provides a mock function with given fields:
func (_m *Application) GetExternalInitiatorManager() webhook.ExternalInitiatorManager {
	ret := _m.Called()

	var r0 webhook.ExternalInitiatorManager
	if rf, ok := ret.Get(0).(func() webhook.ExternalInitiatorManager); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(webhook.ExternalInitiatorManager)
		}
	}

	return r0
}

// GetFeedsService provides a mock function with given fields:
func (_m *Application) GetFeedsService() feeds.Service {
	ret := _m.Called()

	var r0 feeds.Service
	if rf, ok := ret.Get(0).(func() feeds.Service); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(feeds.Service)
		}
	}

	return r0
}

// GetHeadBroadcaster provides a mock function with given fields:
func (_m *Application) GetHeadBroadcaster() types.HeadBroadcasterRegistry {
	ret := _m.Called()

	var r0 types.HeadBroadcasterRegistry
	if rf, ok := ret.Get(0).(func() types.HeadBroadcasterRegistry); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.HeadBroadcasterRegistry)
		}
	}

	return r0
}

// GetHealthChecker provides a mock function with given fields:
func (_m *Application) GetHealthChecker() health.Checker {
	ret := _m.Called()

	var r0 health.Checker
	if rf, ok := ret.Get(0).(func() health.Checker); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(health.Checker)
		}
	}

	return r0
}

// GetKeyStore provides a mock function with given fields:
func (_m *Application) GetKeyStore() *keystore.Master {
	ret := _m.Called()

	var r0 *keystore.Master
	if rf, ok := ret.Get(0).(func() *keystore.Master); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*keystore.Master)
		}
	}

	return r0
}

// GetLogger provides a mock function with given fields:
func (_m *Application) GetLogger() *logger.Logger {
	ret := _m.Called()

	var r0 *logger.Logger
	if rf, ok := ret.Get(0).(func() *logger.Logger); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*logger.Logger)
		}
	}

	return r0
}

// GetStore provides a mock function with given fields:
func (_m *Application) GetStore() *store.Store {
	ret := _m.Called()

	var r0 *store.Store
	if rf, ok := ret.Get(0).(func() *store.Store); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*store.Store)
		}
	}

	return r0
}

// JobORM provides a mock function with given fields:
func (_m *Application) JobORM() job.ORM {
	ret := _m.Called()

	var r0 job.ORM
	if rf, ok := ret.Get(0).(func() job.ORM); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(job.ORM)
		}
	}

	return r0
}

// JobSpawner provides a mock function with given fields:
func (_m *Application) JobSpawner() job.Spawner {
	ret := _m.Called()

	var r0 job.Spawner
	if rf, ok := ret.Get(0).(func() job.Spawner); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(job.Spawner)
		}
	}

	return r0
}

// NewBox provides a mock function with given fields:
func (_m *Application) NewBox() packr.Box {
	ret := _m.Called()

	var r0 packr.Box
	if rf, ok := ret.Get(0).(func() packr.Box); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(packr.Box)
	}

	return r0
}

// PipelineORM provides a mock function with given fields:
func (_m *Application) PipelineORM() pipeline.ORM {
	ret := _m.Called()

	var r0 pipeline.ORM
	if rf, ok := ret.Get(0).(func() pipeline.ORM); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pipeline.ORM)
		}
	}

	return r0
}

// ReplayFromBlock provides a mock function with given fields: number
func (_m *Application) ReplayFromBlock(number uint64) error {
	ret := _m.Called(number)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint64) error); ok {
		r0 = rf(number)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ResumeJobV2 provides a mock function with given fields: ctx, run
func (_m *Application) ResumeJobV2(ctx context.Context, run *pipeline.Run) (bool, error) {
	ret := _m.Called(ctx, run)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, *pipeline.Run) bool); ok {
		r0 = rf(ctx, run)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *pipeline.Run) error); ok {
		r1 = rf(ctx, run)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RunJobV2 provides a mock function with given fields: ctx, jobID, meta
func (_m *Application) RunJobV2(ctx context.Context, jobID int32, meta map[string]interface{}) (int64, error) {
	ret := _m.Called(ctx, jobID, meta)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, int32, map[string]interface{}) int64); ok {
		r0 = rf(ctx, jobID, meta)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int32, map[string]interface{}) error); ok {
		r1 = rf(ctx, jobID, meta)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RunWebhookJobV2 provides a mock function with given fields: ctx, jobUUID, requestBody, meta
func (_m *Application) RunWebhookJobV2(ctx context.Context, jobUUID uuid.UUID, requestBody string, meta pipeline.JSONSerializable) (int64, error) {
	ret := _m.Called(ctx, jobUUID, requestBody, meta)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string, pipeline.JSONSerializable) int64); ok {
		r0 = rf(ctx, jobUUID, requestBody, meta)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, string, pipeline.JSONSerializable) error); ok {
		r1 = rf(ctx, jobUUID, requestBody, meta)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetServiceLogger provides a mock function with given fields: ctx, service, level
func (_m *Application) SetServiceLogger(ctx context.Context, service string, level zapcore.Level) error {
	ret := _m.Called(ctx, service, level)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, zapcore.Level) error); ok {
		r0 = rf(ctx, service, level)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Start provides a mock function with given fields:
func (_m *Application) Start() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Stop provides a mock function with given fields:
func (_m *Application) Stop() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WakeSessionReaper provides a mock function with given fields:
func (_m *Application) WakeSessionReaper() {
	_m.Called()
}
