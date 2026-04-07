// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	bknsdk "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
)

// UploadBKN 上传 BKN tar 包并导入（外部接口）
func (r *restHandler) UploadBKN(c *gin.Context) {
	logger.Debug("Handler UploadBKN Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"上传BKN并导入", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_InvalidParameter).
			WithErrorDetails("Failed to get uploaded file: " + err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	defer file.Close()

	// 验证文件类型
	if header.Header.Get("Content-Type") != "application/octet-stream" {
		// 尝试通过后缀名判断
		ext := filepath.Ext(header.Filename)
		if ext != ".tar" && ext != ".tgz" && ext != ".tar.gz" {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_InvalidParameter).
				WithErrorDetails("Invalid file type, expected tar archive")
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// 获取表单参数
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)

	// 从header中获取业务域（可选）
	businessDomain := c.GetHeader(interfaces.HTTP_HEADER_BUSINESS_DOMAIN)

	logger.Debugf("Upload BKN: branch=%s, filename=%s, size=%d",
		branch, header.Filename, header.Size)

	// 直接从 tar 包加载网络（纯内存，无需写入磁盘）
	bknNetwork, err := bknsdk.LoadNetworkFromTar(file)
	if err != nil {
		logger.Errorf("Failed to load network from tar: %s", err.Error())
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_InvalidParameter).
			WithErrorDetails("Failed to load network from tar: " + err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	bknNetwork.Branch = branch
	bknNetwork.BusinessDomain = businessDomain

	// 执行导入
	kn := logics.ToADPNetWork(bknNetwork)
	otMap := make(map[string]*interfaces.ObjectType)
	for _, bknObj := range bknNetwork.ObjectTypes {
		ot := logics.ToADPObjectType(kn.KNID, kn.Branch, bknObj)
		kn.ObjectTypes = append(kn.ObjectTypes, ot)
		otMap[ot.OTID] = ot
	}
	for _, bknRel := range bknNetwork.RelationTypes {
		rt := logics.ToADPRelationType(kn.KNID, kn.Branch, bknRel)
		kn.RelationTypes = append(kn.RelationTypes, rt)
	}
	for _, bknAct := range bknNetwork.ActionTypes {
		act := logics.ToADPActionType(kn.KNID, kn.Branch, bknAct)
		kn.ActionTypes = append(kn.ActionTypes, act)
	}
	for _, bknCG := range bknNetwork.ConceptGroups {
		cg := logics.ToADPConceptGroup(kn.KNID, kn.Branch, bknCG)
		kn.ConceptGroups = append(kn.ConceptGroups, cg)

		for _, otID := range bknCG.ObjectTypes {
			if ot, ok := otMap[otID]; ok {
				ot.ConceptGroups = append(ot.ConceptGroups, cg)
			}
		}
	}

	// 1. 校验 业务知识网络必要创建参数的合法性, 非空、长度、是枚举值
	err = ValidateKN(ctx, kn)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Validate knowledge network[%s] failed: %s. %v", kn.KNName,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		span.SetAttributes(attr.Key("kn_name").String(kn.KNName))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 若kn的对象类，关系类，行动类, 概念分组不为空，则应循环调用对象类、关系类、行动类, 概念分组的校验函数
	if len(kn.ObjectTypes) > 0 {
		err = ValidateObjectTypes(ctx, kn.KNID, kn.ObjectTypes)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}
	if len(kn.RelationTypes) > 0 {
		err = ValidateRelationTypes(ctx, kn.KNID, kn.RelationTypes)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}
	if len(kn.ActionTypes) > 0 {
		err = ValidateActionTypes(ctx, kn.KNID, kn.ActionTypes)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}
	if len(kn.ConceptGroups) > 0 {
		for _, conceptGroup := range kn.ConceptGroups {
			err = ValidateConceptGroup(ctx, conceptGroup)
			if err != nil {
				httpErr := err.(*rest.HTTPError)
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}
	if len(kn.RiskTypes) > 0 {
		err = ValidateRiskTypes(ctx, kn.KNID, kn.RiskTypes)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// 调用创建单个知识网络
	knID, err := r.kns.CreateKN(ctx, kn, interfaces.ImportMode_Overwrite, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功创建记录审计日志
	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		interfaces.GenerateKNAuditObject(knID, kn.KNName), "")

	logger.Debugf("Upload BKN completed: kn_id=%s", knID)
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, map[string]string{"kn_id": knID})
}

// DownloadBKN 下载 BKN tar 包（外部接口）
func (r *restHandler) DownloadBKN(c *gin.Context) {
	logger.Debug("Handler DownloadBKN Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"下载BKN包", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 获取路径参数
	kn_id := c.Param("kn_id")
	if kn_id == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_InvalidParameter).
			WithErrorDetails("kn_id is required")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 获取查询参数
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)

	logger.Debugf("Download BKN: kn_id=%s, branch=%s", kn_id, branch)

	// 调用服务导出为 tar 包
	tarData, err := r.bs.ExportToTar(ctx, kn_id, branch)
	if err != nil {
		logger.Errorf("Download BKN failed: %s", err.Error())
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_KnowledgeNetwork_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	filename := kn_id + "-" + branch + ".tar"

	logger.Debugf("Download BKN completed: filename=%s size=%d", filename, len(tarData))

	// 设置响应头
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/octet-stream", tarData)
}
