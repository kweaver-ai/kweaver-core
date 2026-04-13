package agentsvc

import (
	"context"
	"fmt"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/static"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

func (agentSvc *agentSvc) GetAPIDoc(ctx context.Context, req *agentreq.GetAPIDocReq) (interface{}, error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()

	oteltrace.SetAttributes(ctx, attribute.String("agent_id", req.AgentID))
	oteltrace.SetAttributes(ctx, attribute.String("agent_version", req.AgentVersion))

	agentInfo, err := agentSvc.squareSvc.GetAgentInfoByIDOrKey(ctx, &squarereq.AgentInfoReq{
		AgentID:      req.AgentID,
		AgentVersion: req.AgentVersion,
	})
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[GetAPIDoc] get agent failed: %v", err), err)
		return nil, errors.Wrapf(err, "[GetAPIDoc] get agent failed: %v", err)
	}

	loader := openapi3.NewLoader()

	docByte, err := static.StaticFiles.ReadFile("agent-api.yaml")
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[GetAPIDoc] read file failed: %v", err), err)
		return nil, errors.Wrapf(err, "[GetAPIDoc] read file failed: %v", err)
	}

	apiDoc, err := loader.LoadFromData(docByte)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[GetAPIDoc] load api doc err: %v", err), err)
		return nil, errors.Wrapf(err, "[GetAPIDoc] load api doc err: %v", err)
	}

	// 3. 取这个接口的配置
	pathItem := apiDoc.Paths.Value("/api/agent-factory/v1/api/chat/completion")
	pathItem.Post.Summary = agentInfo.DataAgent.Name
	ensureBizDomainHeaderParameter(pathItem.Post)

	profile := agentInfo.DataAgent.GetProfileStr()
	if profile != "" {
		pathItem.Post.Description = profile
	}

	// 4. 取请求体
	chatRequest := apiDoc.Components.Schemas["ChatRequest"]
	// 初始化示例
	reqExample := make(map[string]interface{})
	reqExample["custom_querys"] = make(map[string]interface{})
	excludeFields := []string{"history", "query", "header", "tool", "self_config"}

	for _, input := range agentInfo.Config.Input.Fields {
		if input.Name == "history" {
			reqExample[input.Name] = []map[string]string{}
		} else {
			if input.Name == "query" {
				// query的默认值
				reqExample[input.Name] = "在这里输入的问题"
				chatRequest.Value.Properties[input.Name] = &openapi3.SchemaRef{
					Ref: "",
					Value: &openapi3.Schema{
						Type:        &openapi3.Types{openapi3.TypeString},
						Description: "用户输入的问题",
					},
				}
			} else {
				if input.Type == "object" {
					if slices.Contains(excludeFields, input.Name) {
						continue
					}
					// 空结构体
					reqExample["custom_querys"].(map[string]interface{})[input.Name] = make(map[string]interface{})
					chatRequest.Value.Properties[input.Name] = &openapi3.SchemaRef{
						Ref: "",
						Value: &openapi3.Schema{
							Type:        &openapi3.Types{openapi3.TypeObject},
							Description: "用户自定义的输入参数",
						},
					}
				} else if input.Type == "string" {
					if slices.Contains(excludeFields, input.Name) {
						continue
					}

					reqExample["custom_querys"].(map[string]interface{})[input.Name] = "在这里输入参数内容"
					chatRequest.Value.Properties[input.Name] = &openapi3.SchemaRef{
						Ref: "",
						Value: &openapi3.Schema{
							Type:        &openapi3.Types{openapi3.TypeString},
							Description: "用户自定义的输入参数",
						},
					}
				}
			}
		}
	}

	if len(reqExample["custom_querys"].(map[string]interface{})) == 0 {
		delete(reqExample, "custom_querys")
	}

	// 5. 设置请求体示例和schema
	reqExample["stream"] = false
	reqExample["agent_version"] = agentInfo.Version
	reqExample["agent_key"] = agentInfo.DataAgent.Key
	pathItem.Post.RequestBody.Value.Content["application/json"].Example = reqExample
	pathItem.Post.RequestBody.Value.Content["application/json"].Schema = chatRequest

	// 6. 返回
	return apiDoc, nil
}

func ensureBizDomainHeaderParameter(operation *openapi3.Operation) {
	if operation == nil {
		return
	}

	headerName := cenum.HeaderXBizDomainID.String()
	required := !global.GConfig.IsBizDomainDisabled()

	for _, parameter := range operation.Parameters {
		if parameter == nil || parameter.Value == nil || parameter.Value.Name != headerName {
			continue
		}

		parameter.Value.In = "header"
		parameter.Value.Required = required
		parameter.Value.Description = "业务域标识Header，用于区分不同业务场景的请求"
		parameter.Value.Example = "bd_public"
		parameter.Value.Schema = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: &openapi3.Types{openapi3.TypeString},
			},
		}

		return
	}

	operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			In:          "header",
			Name:        headerName,
			Required:    required,
			Description: "业务域标识Header，用于区分不同业务场景的请求",
			Example:     "bd_public",
			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{openapi3.TypeString},
				},
			},
		},
	})
}
