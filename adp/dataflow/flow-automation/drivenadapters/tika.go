// package drivenadapters 旧版纯文本提取器，依赖tika、fasttext
package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/tika.go -destination ../tests/mock_drivenadapters/tika_mock.go

// Tika method interface
type Tika interface {
	// 转换内容
	ParseContent(ctx context.Context, filename string, con *[]byte) (*[]byte, error)

	// 解析元数据
	ParseMetadata(ctx context.Context, filename string, con *[]byte) (*Metadata, error)

	// 匹配文本内容
	MatchContent(ctx context.Context, con *[]byte, tpl map[string]interface{}) (*MatchResult, error)

	// 检查FastTextAnalysys服务是否可用
	CheckFastTextAnalysys(ctx context.Context) error

	// GetPrivacyTemplate 获取隐私识别模板
	GetPrivacyTemplate(ctx context.Context) ([]string, error)
}

type tika struct {
	tikaURL         string
	TextAnalysisURL string
	httpClient      otelHttp.HTTPClient
}

// MatchResult 匹配结果
type MatchResult struct {
	HasPrivateInfo bool                   `json:"has_privacy_info"`
	Results        map[string]MatchDetail `json:"results"`
}

// MatchDetail 匹配详情
type MatchDetail struct {
	Hit  bool        `json:"hit"`
	Info []MatchInfo `json:"match_info"`
}

// MatchInfo 匹配信息
type MatchInfo struct {
	Content  string  `json:"content"`
	Position []int   `json:"position"`
	Prob     float64 `json:"prob"`
}

// Metadata 文档元数据
type Metadata struct {
	Page string `json:"xmpTPg:NPages"`
}

var (
	tOnce sync.Once
	tk    Tika
)

// NewTika 创建tika服务
func NewTika() Tika {
	tOnce.Do(func() {
		config := common.NewConfig()
		tk = &tika{
			tikaURL:         fmt.Sprintf("http://%s:%v", config.Tika.Host, config.Tika.Port),
			TextAnalysisURL: fmt.Sprintf("http://%s:%v", config.FastTextAnalysis.Host, config.FastTextAnalysis.Port),
			httpClient:      NewOtelHTTPClient(),
		}
	})
	return tk
}

// 转换内容
func (t *tika) ParseContent(ctx context.Context, filename string, con *[]byte) (*[]byte, error) {
	target := fmt.Sprintf("%s/tika", t.tikaURL)
	headers := map[string]string{
		"Content-Disposition": fmt.Sprintf("attachment; filename=%s", filename),
		"Accept":              "text/plain",
	}
	_, body, err := t.httpClient.Request(ctx, target, http.MethodPut, headers, con)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("ParseContent failed: %v, url: %v", err, target)
		return nil, err
	}

	return &body, nil
}

// ParseMetadata 解析元数据
func (t *tika) ParseMetadata(ctx context.Context, filename string, con *[]byte) (*Metadata, error) {
	log := traceLog.WithContext(ctx)
	target := fmt.Sprintf("%s/meta", t.tikaURL)
	headers := map[string]string{
		"Content-Disposition": fmt.Sprintf("attachment; filename=%s", filename),
		"Accept":              "application/json",
	}
	_, body, err := t.httpClient.Request(ctx, target, http.MethodPut, headers, con)
	if err != nil {
		log.Warnf("ParseContent failed: %v, url: %v", err, target)
		return nil, err
	}
	var metadata Metadata
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		log.Warnf("ParseContent Unmarshal failed: %v, body: %v", err, body)
	}

	return &metadata, nil
}

// 检查FastTextAnalysys服务是否可用
func (t *tika) CheckFastTextAnalysys(ctx context.Context) error {
	target := fmt.Sprintf("%s/health/ready", t.TextAnalysisURL)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, _, err := t.httpClient.Get(ctx, target, map[string]string{})

	return err
}

// MatchContent 匹配文本内容
func (t *tika) MatchContent(ctx context.Context, con *[]byte, tpl map[string]interface{}) (*MatchResult, error) {
	log := traceLog.WithContext(ctx)
	target := fmt.Sprintf("%s/api/fast-text-analysis/v1/recognize-privacy", t.TextAnalysisURL)
	body := map[string]interface{}{
		"content": string(*con),
		"level":   3,
	}

	templates := []map[string]interface{}{tpl}

	if _, ok := tpl["name"]; ok {
		body["templates"] = templates
	}

	_, respParam, err := t.httpClient.Post(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, body)

	if err != nil {
		log.Warnf("MatchContent failed: %v, url: %v", err, target)
		return nil, err
	}

	resByte, err := json.Marshal(respParam)
	if err != nil {
		log.Warnf("MatchContent parse data failed: %v, url: %v", err, target)
		return nil, err
	}

	var matchResult MatchResult
	err = json.Unmarshal(resByte, &matchResult)
	if err != nil {
		log.Warnf("MatchContent parse data failed: %v, url: %v", err, target)
		return nil, err
	}

	return &matchResult, nil
}

// GetPrivacyTemplate 获取隐私识别模板
func (t *tika) GetPrivacyTemplate(ctx context.Context) ([]string, error) {
	log := traceLog.WithContext(ctx)
	target := fmt.Sprintf("%s/api/fast-text-analysis/v1/recognize-privacy/builtin", t.TextAnalysisURL)
	newCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, respParam, err := t.httpClient.Get(newCtx, target, map[string]string{})
	if err != nil {
		log.Warnf("GetPrivacyTemplate failed: %v, url: %v", err, target)
		return nil, err
	}

	resByte, err := json.Marshal(respParam)
	if err != nil {
		log.Warnf("GetPrivacyTemplate parse data failed: %v, url: %v", err, target)
		return nil, err
	}

	var res []string
	err = json.Unmarshal(resByte, &res)
	if err != nil {
		log.Warnf("GetPrivacyTemplate parse data failed: %v, url: %v", err, target)
		return nil, err
	}

	return res, err
}
