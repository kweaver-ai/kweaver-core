package smoke

import (
	"context"
	"fmt"
	"net/url"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

// 冒烟测试用例
type smokeClient struct {
	baseURL string
	client  interfaces.HTTPClient
}

// NewPublicSmokeClient 创建一个新的测试客户端
func NewPublicSmokeClient(host string) *smokeClient {
	baseURL := fmt.Sprintf("http://%s/api/agent-operator-integration/v1", host)
	return &smokeClient{
		baseURL: baseURL,
		client:  rest.NewHTTPClient(),
	}
}

func NewPrivateSmokeClient(host string) *smokeClient {
	baseURL := fmt.Sprintf("http://%s/api/agent-operator-integration/internal-v1", host)
	return &smokeClient{
		baseURL: baseURL,
		client:  rest.NewHTTPClient(),
	}
}

// Register 注册
func (s *smokeClient) Register(ctx context.Context, req *interfaces.OperatorRegisterReq, token string) (int, interface{}, error) {
	url := s.baseURL + "/operator/register"
	header := map[string]string{}
	if token != "" {
		header["Authorization"] = fmt.Sprintf("Bearer %s", token)
	}
	code, result, err := s.client.Post(ctx, url, header, req)
	if err != nil {
		return code, nil, err
	}
	fmt.Println(code)
	return code, result, nil
}

func (s *smokeClient) GetList(ctx context.Context, token string) (interface{}, error) {
	src := s.baseURL + "/operator/info/list"
	queryValues := url.Values{}
	queryValues.Add("page", "1")
	queryValues.Add("page_size", "10")
	header := map[string]string{}
	if token != "" {
		header["Authorization"] = fmt.Sprintf("Bearer %s", token)
	}
	code, result, err := s.client.Get(ctx, src, queryValues, header)
	if err != nil {
		return nil, err
	}
	fmt.Println(code)
	return result, nil
}

func (s *smokeClient) GetInfo(ctx context.Context, operatorID, version, token string) (interface{}, error) {
	src := s.baseURL + fmt.Sprintf("/operator/info/%s", operatorID)
	header := map[string]string{}
	if token != "" {
		header["Authorization"] = fmt.Sprintf("Bearer %s", token)
	}
	var queryValues url.Values
	if version != "" {
		queryValues = url.Values{}
		queryValues.Add("version", version)
	}
	code, result, err := s.client.Get(ctx, src, queryValues, header)
	if err != nil {
		return nil, err
	}
	fmt.Println(code)
	return result, nil
}
