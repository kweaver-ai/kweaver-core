//go:build !skiptest
// +build !skiptest

package local

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
// 	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
// 	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
// )

// func TestCheckIsExistByName(t *testing.T) {
// 	odb := dbaccess.NewOperatorManagerDB()

// 	exist, opID, version, err := odb.RegisterOperator(context.Background(), &model.OperatorRegisterDB{
// 		OperatorID:     "123",
// 		Name:           "test",
// 		MetadataType:   "apis",
// 		Status:         string(interfaces.OperatorStatusUnpublished),
// 		OperatorType:   string(interfaces.OperatorTypeBase),
// 		ExecutionMode:  string(interfaces.OperatorExecutionModeSync),
// 		Category:       "{}",
// 		ExecuteControl: "test",
// 		ExtendInfo:     "test",
// 		CreateUser:     "test",
// 		CreateTime:     time.Now().UnixNano(),
// 		UpdateUser:     "test",
// 		UpdateTime:     time.Now().UnixNano(),
// 	}, &model.APIMetadataDB{
// 		Summary:     "test",
// 		Version:     "1.0.0",
// 		Description: "test",
// 		Path:        "/test",
// 		ServerURL:  "http://test.com",
// 		Method:      "GET",
// 		APISpec:     "test",
// 		CreateUser:  "test",
// 		CreateTime:  time.Now().UnixNano(),
// 		UpdateUser:  "test",
// 		UpdateTime:  time.Now().UnixNano(),
// 	}, false)

// 	if err != nil {
// 		t.Fatalf("err: %v", err)
// 	}
// 	t.Logf("exist: %v, opID: %v, version: %v", exist, opID, version)
// }
