package mongo

// import (
// 	"fmt"
// 	"testing"
// 	"time"

// 	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity/entity"
// 	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
// 	"github.com/stretchr/testify/assert"
// )

// var mongoConn = "mongodb://root:pwd@127.0.0.1:27017/automation?authSource=admin"

// func TestStore_Dag(t *testing.T) {
// 	s := NewStore(&StoreOption{
// 		ConnStr: mongoConn,
// 	})

// 	err := s.Init()
// 	assert.NoError(t, err)
// 	giveDag := []*entity.Dag{
// 		{
// 			BaseInfo: entity.BaseInfo{
// 				ID: "test1",
// 			},
// 			Tasks: []entity.Task{{ID: "test"}},
// 		},
// 		{
// 			BaseInfo: entity.BaseInfo{
// 				ID: "test2",
// 			},
// 			Tasks: []entity.Task{{ID: "test"}},
// 		},
// 	}
// 	// create
// 	for i := range giveDag {
// 		_, err := s.CreateDag(giveDag[i])
// 		assert.NoError(t, err)
// 	}

// 	ret, err := s.ListDag(nil)
// 	assert.NoError(t, err)
// 	time.Sleep(time.Second)
// 	// check and update
// 	for i := range ret {
// 		assert.NotEqual(t, "", ret[i].ID)
// 		assert.Greater(t, ret[i].CreatedAt, int64(0))
// 		assert.Greater(t, ret[i].UpdatedAt, int64(0))
// 		ret[i].Name = fmt.Sprintf("name-%d", i)
// 		ret[i].Desc = fmt.Sprintf("desc-%d", i)
// 		ret[i].Cron = fmt.Sprintf("cron-%d", i)
// 		ret[i].Vars = entity.DagVars{
// 			"var1": {}, "var2": {},
// 		}
// 		ret[i].Tasks = []entity.Task{
// 			{ID: "task1", Name: "task1"}, {ID: "task2", Name: "task2"},
// 		}

// 		err = s.UpdateDag(ret[i])
// 		assert.NoError(t, err)
// 	}

// 	ret, err = s.ListDag(nil)
// 	assert.NoError(t, err)
// 	for i := range ret {
// 		assert.NotEmpty(t, ret[i].ID)
// 		assert.NotEmpty(t, ret[i].Name)
// 		assert.NotEmpty(t, ret[i].Desc)
// 		assert.NotEmpty(t, ret[i].Cron)
// 		assert.NotEqual(t, ret[i].CreatedAt, ret[i].UpdatedAt)
// 	}

// 	// delete
// 	err = s.BatchDeleteDag([]string{"test1", "test2"})
// 	assert.NoError(t, err)
// 	ret, err = s.ListDag(nil)
// 	assert.NoError(t, err)
// 	assert.Equal(t, 0, len(ret))
// }

// func TestStore_DagIns(t *testing.T) {
// 	s := NewStore(&StoreOption{
// 		ConnStr: mongoConn,
// 	})
// 	entity.StoreMarshal = s.Marshal
// 	entity.StoreUnmarshal = s.Unmarshal

// 	err := s.Init()
// 	assert.NoError(t, err)
// 	giveDagIns := []*entity.DagInstance{
// 		{
// 			BaseInfo: entity.BaseInfo{
// 				ID: "test1",
// 			},
// 		},
// 		{
// 			BaseInfo: entity.BaseInfo{
// 				ID: "test2",
// 			},
// 		},
// 	}
// 	// create
// 	for i := range giveDagIns {
// 		err := s.CreateDagIns(giveDagIns[i])
// 		assert.NoError(t, err)
// 	}

// 	ret, err := s.ListDagInstance(&mod.ListDagInstanceInput{})
// 	assert.NoError(t, err)
// 	time.Sleep(time.Second)
// 	// check and update
// 	for i := range ret {
// 		assert.NotEqual(t, "", ret[i].ID)
// 		assert.Greater(t, ret[i].CreatedAt, int64(0))
// 		assert.Greater(t, ret[i].UpdatedAt, int64(0))
// 		ret[i].Worker = fmt.Sprintf("worker-%d", i)
// 		ret[i].DagID = fmt.Sprintf("dagid-%d", i)
// 		ret[i].ShareData = &entity.ShareData{Dict: map[string]interface{}{
// 			"test": "gg",
// 		}}
// 		ret[i].Vars = entity.DagInstanceVars{
// 			"var1": {}, "var2": {},
// 		}

// 		err = s.UpdateDagIns(ret[i])
// 		assert.NoError(t, err)
// 	}

// 	ret, err = s.ListDagInstance(&mod.ListDagInstanceInput{})
// 	assert.NoError(t, err)
// 	for i := range ret {
// 		assert.NotEmpty(t, ret[i].ID)
// 		assert.NotEmpty(t, ret[i].Worker)
// 		assert.NotEmpty(t, ret[i].DagID)
// 		assert.NotNil(t, ret[i].ShareData)
// 		assert.NotEqual(t, ret[i].CreatedAt, ret[i].UpdatedAt)
// 	}

// 	// delete
// 	err = s.BatchDeleteDagIns([]string{"test1", "test2"})
// 	assert.NoError(t, err)
// 	ret, err = s.ListDagInstance(&mod.ListDagInstanceInput{})
// 	assert.NoError(t, err)
// 	assert.Equal(t, 0, len(ret))
// }

// func TestStore_TaskIns(t *testing.T) {
// 	s := NewStore(&StoreOption{
// 		ConnStr: mongoConn,
// 	})

// 	err := s.Init()
// 	assert.NoError(t, err)
// 	giveTaskIns := []*entity.TaskInstance{
// 		{
// 			BaseInfo: entity.BaseInfo{
// 				ID: "test1",
// 			},
// 		},
// 		{
// 			BaseInfo: entity.BaseInfo{
// 				ID: "test2",
// 			},
// 		},
// 	}
// 	// create
// 	err = s.BatchCreatTaskIns(giveTaskIns)
// 	assert.NoError(t, err)

// 	ret, err := s.ListTaskInstance(&mod.ListTaskInstanceInput{})
// 	assert.NoError(t, err)
// 	time.Sleep(time.Second)
// 	// check and update
// 	for i := range ret {
// 		assert.NotEqual(t, "", ret[i].ID)
// 		assert.Greater(t, ret[i].CreatedAt, int64(0))
// 		assert.Greater(t, ret[i].UpdatedAt, int64(0))
// 		ret[i].Name = fmt.Sprintf("name-%d", i)
// 		ret[i].DagInsID = fmt.Sprintf("dag-%d", i)
// 		ret[i].ActionName = fmt.Sprintf("act-%d", i)
// 		ret[i].DependOn = []string{"test1"}

// 		err = s.UpdateTaskIns(ret[i])
// 		assert.NoError(t, err)
// 	}

// 	ret, err = s.ListTaskInstance(&mod.ListTaskInstanceInput{})
// 	assert.NoError(t, err)
// 	for i := range ret {
// 		assert.NotEmpty(t, ret[i].ID)
// 		assert.NotEmpty(t, ret[i].DagInsID)
// 		assert.NotEmpty(t, ret[i].ActionName)
// 		assert.NotNil(t, ret[i].DependOn)
// 	}

// 	// delete
// 	err = s.BatchDeleteTaskIns([]string{"test1", "test2"})
// 	assert.NoError(t, err)
// 	ret, err = s.ListTaskInstance(&mod.ListTaskInstanceInput{})
// 	assert.NoError(t, err)
// 	assert.Equal(t, 0, len(ret))
// }
