package agentinoutsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/daconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
)

const (
	MaxFileSize = 10 * 1024 * 1024 // 10MB
)

// Import 导入agent数据
func (s *agentInOutSvc) Import(ctx context.Context, req *agentinoutreq.ImportReq) (resp *agentinoutresp.ImportResp, err error) {
	resp = agentinoutresp.NewImportResp()

	// 1. 获取用户ID
	uid := chelper.GetUserIDFromCtx(ctx)
	if uid == "" {
		err = capierr.New400Err(ctx, "无法获取用户ID")
		return
	}

	// 2. 检查文件大小
	if req.File.Size > MaxFileSize {
		err = capierr.New400Err(ctx, "文件大小不能超过10MB")
		return
	}

	// 3. 检查文件类型
	if req.File.Header.Get("Content-Type") != "application/json" {
		err = capierr.New400Err(ctx, "只支持JSON格式文件")
		return
	}

	// 4. 打开文件
	file, err := req.File.Open()
	if err != nil {
		err = capierr.NewCustom400Err(ctx, apierr.AgentFactoryInoutParseFileFailed, "无法打开上传文件")
		return
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	// 5. 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		err = capierr.NewCustom400Err(ctx, apierr.AgentFactoryInoutParseFileFailed, "无法读取文件内容")
		return
	}

	// 6. 解析JSON
	var exportData agentinoutresp.ExportResp
	if err = json.Unmarshal(content, &exportData); err != nil {
		err = capierr.NewCustom400Err(ctx, apierr.AgentFactoryInoutParseFileFailed, "文件格式错误，无法解析JSON")
		return
	}

	if len(exportData.Agents) == 0 {
		err = capierr.NewCustom400Err(ctx, apierr.AgentFactoryInoutParseFileFailed, "导入文件中没有agent数据")
		return
	}

	// 6.1 校验单次导入最多导入xx个agent
	maxSize := daconstant.AgentInoutMaxSize
	if len(exportData.Agents) > maxSize {
		err = capierr.NewCustom400Err(ctx, apierr.AgentFactoryInoutMaxSizeExceeded, fmt.Sprintf("单次导入最多导入%d个agent", maxSize))
		return
	}

	// 7. 公共检查
	err = s.importCheck(ctx, &exportData, resp)
	if err != nil {
		return
	}

	if resp.HasFail() {
		return
	}

	// 8. 导入

	if req.ImportType == agentinoutreq.ImportTypeCreate {
		err = s.importByCreate(ctx, &exportData, uid, resp)
	} else if req.ImportType == agentinoutreq.ImportTypeUpsert {
		err = s.importByUpsert(ctx, &exportData, uid, resp)
	}

	if err != nil {
		return
	}

	// 8.1 检查导入结果 如果失败则返回
	if resp.HasFail() {
		return
	}

	// 8.2 设置成功
	resp.IsSuccess = true

	return
}
