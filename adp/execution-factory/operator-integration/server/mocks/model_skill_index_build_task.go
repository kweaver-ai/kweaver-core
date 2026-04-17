// Code generated manually for ISkillIndexBuildTaskDB. DO NOT EDIT lightly.
package mocks

import (
	context "context"
	sql "database/sql"
	reflect "reflect"

	ormhelper "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	model "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	gomock "go.uber.org/mock/gomock"
)

type MockISkillIndexBuildTaskDB struct {
	ctrl     *gomock.Controller
	recorder *MockISkillIndexBuildTaskDBMockRecorder
	isgomock struct{}
}

type MockISkillIndexBuildTaskDBMockRecorder struct {
	mock *MockISkillIndexBuildTaskDB
}

func NewMockISkillIndexBuildTaskDB(ctrl *gomock.Controller) *MockISkillIndexBuildTaskDB {
	mock := &MockISkillIndexBuildTaskDB{ctrl: ctrl}
	mock.recorder = &MockISkillIndexBuildTaskDBMockRecorder{mock}
	return mock
}

func (m *MockISkillIndexBuildTaskDB) EXPECT() *MockISkillIndexBuildTaskDBMockRecorder {
	return m.recorder
}

func (m *MockISkillIndexBuildTaskDB) Insert(ctx context.Context, tx *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", ctx, tx, task)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockISkillIndexBuildTaskDBMockRecorder) Insert(ctx, tx, task any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockISkillIndexBuildTaskDB)(nil).Insert), ctx, tx, task)
}

func (m *MockISkillIndexBuildTaskDB) SelectByTaskID(ctx context.Context, tx *sql.Tx, taskID string) (*model.SkillIndexBuildTaskDB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectByTaskID", ctx, tx, taskID)
	ret0, _ := ret[0].(*model.SkillIndexBuildTaskDB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockISkillIndexBuildTaskDBMockRecorder) SelectByTaskID(ctx, tx, taskID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectByTaskID", reflect.TypeOf((*MockISkillIndexBuildTaskDB)(nil).SelectByTaskID), ctx, tx, taskID)
}

func (m *MockISkillIndexBuildTaskDB) SelectRunningTask(ctx context.Context, tx *sql.Tx) (*model.SkillIndexBuildTaskDB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectRunningTask", ctx, tx)
	ret0, _ := ret[0].(*model.SkillIndexBuildTaskDB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockISkillIndexBuildTaskDBMockRecorder) SelectRunningTask(ctx, tx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectRunningTask", reflect.TypeOf((*MockISkillIndexBuildTaskDB)(nil).SelectRunningTask), ctx, tx)
}

func (m *MockISkillIndexBuildTaskDB) SelectLatestCompletedIncrementalTask(ctx context.Context, tx *sql.Tx) (*model.SkillIndexBuildTaskDB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectLatestCompletedIncrementalTask", ctx, tx)
	ret0, _ := ret[0].(*model.SkillIndexBuildTaskDB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockISkillIndexBuildTaskDBMockRecorder) SelectLatestCompletedIncrementalTask(ctx, tx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectLatestCompletedIncrementalTask", reflect.TypeOf((*MockISkillIndexBuildTaskDB)(nil).SelectLatestCompletedIncrementalTask), ctx, tx)
}

func (m *MockISkillIndexBuildTaskDB) SelectListPage(ctx context.Context, tx *sql.Tx, filter map[string]any, sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) ([]*model.SkillIndexBuildTaskDB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectListPage", ctx, tx, filter, sort, cursor)
	ret0, _ := ret[0].([]*model.SkillIndexBuildTaskDB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockISkillIndexBuildTaskDBMockRecorder) SelectListPage(ctx, tx, filter, sort, cursor any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectListPage", reflect.TypeOf((*MockISkillIndexBuildTaskDB)(nil).SelectListPage), ctx, tx, filter, sort, cursor)
}

func (m *MockISkillIndexBuildTaskDB) CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]any) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountByWhereClause", ctx, tx, filter)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockISkillIndexBuildTaskDBMockRecorder) CountByWhereClause(ctx, tx, filter any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountByWhereClause", reflect.TypeOf((*MockISkillIndexBuildTaskDB)(nil).CountByWhereClause), ctx, tx, filter)
}

func (m *MockISkillIndexBuildTaskDB) UpdateByTaskID(ctx context.Context, tx *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateByTaskID", ctx, tx, task)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockISkillIndexBuildTaskDBMockRecorder) UpdateByTaskID(ctx, tx, task any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateByTaskID", reflect.TypeOf((*MockISkillIndexBuildTaskDB)(nil).UpdateByTaskID), ctx, tx, task)
}
