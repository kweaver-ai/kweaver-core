package smoke

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

func TestPublicSmokeOperatorInfo(t *testing.T) {
	client := NewPublicSmokeClient("10.4.175.99")
	ctx := context.Background()
	token := "ory_at_knmVXaatkFwVb7IHO8QRAYuEmRJ0Av9pKshu9agnF_M.0EhBbOEpS5xxQq1AN7a1EIkU7yYXR-q6zYG6yhSRIJc"
	result, err := client.GetInfo(ctx, "f7b7e413-d410-4e0e-b969-820486c213dd", "f9cdb224-9a08-4b19-9f79-cac901f07eb5", token)
	if err != nil {
		fmt.Printf("GetInfo failed %v", err)
		return
	}
	fmt.Println(utils.ObjectToJSON(result))
}

func TestPublicSmokeOperatorList(t *testing.T) {
	client := NewPublicSmokeClient("10.4.110.92")
	ctx := context.Background()
	token := "ory_at_w8OWTKbiO15G6dMYdS5OtC0dNDw-DPmehOx0RSf7C_w.d-WIbKWJru4SZZwXbM25Mr0EaOLPTEwSw4h5aLvIpvM"
	result, err := client.GetList(ctx, token)
	if err != nil {
		fmt.Printf("GetList failed 777 %v", err)
		return
	}
	fmt.Println(utils.ObjectToJSON(result))
}

func TestPublicSmokeOperatorRegister(t *testing.T) {
	client := NewPublicSmokeClient("10.4.175.99")
	data, err := os.ReadFile("../file/yaml/template.yaml")
	if err != nil {
		fmt.Printf("ReadFile failed: %v", err)
		return
	}
	token := "ory_at_21VBfRCRyAJXgLU59g2HFI3lLbrb4CvgaJnjKEQNpBs.FpfYLM4W5KHjaFLbI3UR5GcvI0BUaZXXIsEaM-U6fmc"
	code, result, err := client.Register(context.Background(), &interfaces.OperatorRegisterReq{
		Data:         string(data),
		MetadataType: interfaces.MetadataTypeAPI,
		OperatorInfo: &interfaces.OperatorInfo{
			Category: "other_category",
		},
		DirectPublish: true,
	}, token)
	fmt.Println(code)
	if err != nil {
		fmt.Printf("Register failesssd: =sssss %v", err)
		return
	}
	fmt.Println(code)
	fmt.Println(utils.ObjectToJSON(result))
}

type ExecuteToolReq struct {
	Headers     map[string]string `json:"header"`
	Body        interface{}       `json:"body"`
	QueryParams map[string]string `json:"query"`
	PathParams  map[string]string `json:"path"`
}

func TestPublicSmokeExecuteTool(t *testing.T) {
	local := "/root/go/src/github.com/kweaver-ai/adp/execution-factory/operator-integration/server/tests/file/auth.json"
	client := NewPublicSmokeClient("127.0.0.1:9000")
	data, err := os.ReadFile(local)
	if err != nil {
		fmt.Printf("ReadFile failed: %v", err)
		return
	}
	token := "ory_at_RvKfV0CHK_TCxoa8mXZj5lT0Zw5WSfHe0ot1GuasEBY.FLajysFAPyUyQesRPbIHOROB9GAO1QyE1__57ztdw9U"
	reqBody := map[string]interface{}{
		"metadata_type": "openapi",
		"data":          data,
	}
	reqHeaders := map[string]string{
		"Content-Type": "multipart/form-data",
	}
	reqHeaders["Authorization"] = fmt.Sprintf("Bearer %s", token)
	req := &ExecuteToolReq{
		Body:    reqBody,
		Headers: reqHeaders,
	}
	result, err := client.ExecuteTool(context.Background(), "49d0ac32-0976-4560-a6b8-72dd232b543b", "c2395a45-fdf0-49f8-a996-8d51762753eb", req, token)
	if err != nil {
		fmt.Printf("ExecuteTool failed: %v", err)
		return
	}
	fmt.Println(utils.ObjectToJSON(result))
}
