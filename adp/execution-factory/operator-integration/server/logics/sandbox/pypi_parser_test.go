package sandbox

// import (
// 	"context"
// 	"fmt"
// 	"testing"

// 	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
// )

// func TestParsePypi(t *testing.T) {
// 	resp, err := ParsePypi(context.Background(), &ParsePypiReq{
// 		PypiRepoURL:   DefaultPypiRepo,
// 		PackageName:   "requests",
// 		PythonVersion: "3.10.0",
// 	})
// 	if err != nil {
// 		t.Fatalf("ParsePypi failed: %v", err)
// 	}
// 	fmt.Println(utils.ObjectToJSON(resp))
// }
