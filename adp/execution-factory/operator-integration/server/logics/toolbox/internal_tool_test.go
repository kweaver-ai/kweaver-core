package toolbox

// // 测试工具版本
// func TestXxx(t *testing.T) {
// 	ctx := context.Background()
// 	boxID := "build-in"
// 	s := &ToolServiceImpl{
// 		Validator:     validator.NewValidator(),
// 		OpenAPIParser: parsers.NewOpenAPIParser(),
// 	}
// 	// 读取OpenAPI文档
// 	localPath := "/root/go/src/github.com/kweaver-ai/adp/execution-factory/operator-integration/server/tests/file/yaml/test.yaml"
// 	data, err := os.ReadFile(localPath)
// 	if err != nil {
// 		t.Errorf("ReadFile err: %+v", err)
// 	}
// 	content, err := s.OpenAPIParser.GetAllContent(data)
// 	if err != nil {
// 		t.Errorf("GetAllContent err: %+v", err)
// 	}
// 	// 解析
// 	metadataMap, toolMap, _, _, err := s.parseOpenAPIToMetadata(ctx, boxID, "", content)
// 	if err != nil {
// 		t.Errorf("parseOpenAPIToMetadata err: %+v", err)
// 	}

// 	for key, tool := range toolMap {
// 		ruleProcess, err := NewInternalToolRule(tool, metadataMap[key])
// 		if err != nil {
// 			t.Errorf("NewInternalToolRule err: %+v", err)
// 		}
// 		version, err := ruleProcess.GetVersion()
// 		if err != nil {
// 			t.Errorf("GetVersion err: %+v", err)
// 		}
// 		t.Logf("version: %s", version)
// 	}
// }
