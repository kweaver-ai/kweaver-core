// Package drivenadapters 出站适配器
package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/ecotag.go -destination ../tests/mock_drivenadapters/ecotag_mock.go

// EcoTag method interface
type EcoTag interface {
	GetTags(ctx context.Context, params map[string][]string) ([]*TagInfo, error)
	GetTagTrees(ctx context.Context) ([]*TagTree, error)
}

// TagInfo 标签信息
type TagInfo = common.TagInfo

// TagTree 标签树
type TagTree struct {
	TagInfo
	CreatedAt int64      `json:"created_at"`
	ChildTags []*TagTree `json:"child_tags"`
}

type ecoTag struct {
	baseURL    string
	httpClient otelHttp.HTTPClient
}

var (
	ecoOnce sync.Once
	eco     EcoTag
)

// NewEcoTag new instance
func NewEcoTag() EcoTag {
	ecoOnce.Do(func() {
		config := common.NewConfig()
		eco = &ecoTag{
			baseURL:    fmt.Sprintf("http://%s:%v", config.EcoTag.PrivateHost, config.EcoTag.PrivatePort),
			httpClient: NewOtelHTTPClient(),
		}
	})

	return eco
}

// GetTags get tags
func (e *ecoTag) GetTags(ctx context.Context, params map[string][]string) ([]*TagInfo, error) {
	var query string
	for k, v := range params {
		val := strings.Join(v, ",")
		if strings.EqualFold(k, "path") || strings.EqualFold(k, "id") {
			val = url.QueryEscape(val)
		}
		query = fmt.Sprintf("%s%s=%s&", query, k, val)
	}
	target := fmt.Sprintf("%s/api/ecotag/v1/tag?%s", e.baseURL, strings.TrimSuffix(query, "&"))
	_, respParam, err := e.httpClient.Get(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.GetTags] faild: %v, url: %v", err, target)
		return nil, err
	}

	var tagsInfo = make([]*TagInfo, 0)
	tags := respParam.([]interface{})
	for _, tag := range tags {
		_tag := tag.(map[string]interface{})
		tagsInfo = append(tagsInfo, &TagInfo{
			ID:      _tag["id"].(string),
			Name:    _tag["name"].(string),
			Path:    _tag["path"].(string),
			Version: int(_tag["version"].(float64)),
		})
	}

	return tagsInfo, nil
}

// GetTagTrees get tag tree
func (e *ecoTag) GetTagTrees(ctx context.Context) ([]*TagTree, error) {
	target := fmt.Sprintf("%s/api/ecotag/v1/tag-tree", e.baseURL)
	_, respParam, err := e.httpClient.Get(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.GetTags] faild: %v, url: %v", err, target)
		return nil, err
	}

	var tagTrees = make([]*TagTree, 0)
	respByte, err := json.Marshal(respParam)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.GetTags] Marshal faild, detail: %s", err)
		return nil, err
	}
	err = json.Unmarshal(respByte, &tagTrees)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.GetTags] Unmarshal faild, detail: %s", err)
		return nil, err
	}

	return tagTrees, err
}
