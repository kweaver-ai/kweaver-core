//go:build !skiptest
// +build !skiptest

package smoke

import (
	"context"
	"fmt"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

func TestPrivateSmokeOperatorInfo(t *testing.T) {
	client := NewPrivateSmokeClient("127.0.0.1:9000")
	ctx := context.Background()
	result, err := client.GetInfo(ctx, "f180d69d-9f1b-4f32-b216-99cc5be6cb1a",
		"e09d6839-e8a0-42bc-9c68-61ee5e6b9a15", "")
	if err != nil {
		fmt.Printf("TestPrivateSmokeOperatorInf failessssd000 %v", err)
		return
	}
	fmt.Println(utils.ObjectToJSON(result))
}
func TestPrivateSmokeOperatorList(t *testing.T) {
	client := NewPrivateSmokeClient("127.0.0.1:9000")
	ctx := context.Background()
	token := ""
	result, err := client.GetList(ctx, token)
	if err != nil {
		fmt.Printf("GetList failed %v", err)
		return
	}
	fmt.Println(utils.ObjectToJSON(result))
}
