package smoke

import (
	"context"
	"fmt"
)

// http://127.0.0.1:9000/api/agent-operator-integration/v1/tool-box/{box_id}/proxy/{tool_id}
func (s *smokeClient) ExecuteTool(ctx context.Context, boxID, toolID string, req interface{}, token string) (interface{}, error) {
	// src := s.baseURL + "/toolbox/" + boxID + "/tool/" + toolID + "/execute"
	src := fmt.Sprintf("%s/tool-box/%s/proxy/%s", s.baseURL, boxID, toolID)
	fmt.Println(src)
	header := map[string]string{}
	if token != "" {
		header["Authorization"] = fmt.Sprintf("Bearer %s", token)
	}
	code, result, err := s.client.Post(ctx, src, header, req)
	if err != nil {
		return nil, err
	}
	fmt.Println(code)
	return result, nil
}
