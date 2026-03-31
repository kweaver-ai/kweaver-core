package drivenadapters

import (
	"context"
	"flow-stream-data-pipeline/common"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"flow-stream-data-pipeline/pipeline-worker/interfaces"
)

var (
	ibAccessOnce sync.Once
	ibAccess     interfaces.IndexBaseAccess
)

type indexBaseAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewIndexBaseAccess(appSetting *common.AppSetting) interfaces.IndexBaseAccess {
	ibAccessOnce.Do(func() {
		ibAccess = &indexBaseAccess{
			appSetting: appSetting,
			// 设置超时时间 40s
			httpClient: common.NewHTTPClientWithOptions(rest.HttpClientOptions{TimeOut: 40}),
		}
	})

	return ibAccess
}

// 根据索引库类型获取索引库详情
func (iba *indexBaseAccess) GetIndexBasesByTypes(ctx context.Context, types []string) ([]*interfaces.IndexBaseInfo, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get index bases by types", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	baseTypes := strings.Join(types, ",")
	url := fmt.Sprintf("%s/%s", iba.appSetting.IndexBaseUrl, baseTypes)

	span.SetAttributes(attr.Key("base_types").String(baseTypes))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         url,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, respData, err := iba.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	if err != nil {
		errDetails := fmt.Sprintf("Get indexbases by base types '%s' failed, %s", baseTypes, err.Error())
		logger.Error(errDetails)

		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http GET index bases by types failed")

		return nil, err
	}

	if respCode != http.StatusOK {
		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			errDetails := fmt.Sprintf("Unmalshal baesError failed: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal baesError failed")

			return nil, err
		}

		o11y.Error(ctx, fmt.Sprintf("%s. %v", baseError.Description, baseError.ErrorDetails))
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return nil, fmt.Errorf("get indexbases '%s' failed, errDetails: %v", baseTypes, baseError.ErrorDetails)
	}

	var bases []*interfaces.IndexBaseInfo
	if err := sonic.Unmarshal(respData, &bases); err != nil {
		errDetails := fmt.Sprintf("Unmarshal indexbase respData failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal indexbase info failed")

		return nil, err
	}

	// 索引库接口默认批量查询，只返回存在的索引库信息
	if len(bases) < len(types) {
		nonexistentBases := make([]string, 0)
		typesMap := make(map[string]struct{})

		for _, base := range bases {
			typesMap[base.BaseType] = struct{}{}
		}

		for _, baseType := range types {
			if _, ok := typesMap[baseType]; !ok {
				nonexistentBases = append(nonexistentBases, baseType)
			}
		}

		errDetails := fmt.Sprintf("IndexBases %v doesn't exist", nonexistentBases)
		logger.Warn(errDetails)
		o11y.Warn(ctx, errDetails)
		// 如果有的存在，有的不存在，报错
		return nil, fmt.Errorf("IndexBases %v not found", nonexistentBases)
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return bases, nil
}

func (ibs *indexBaseAccess) GetIndexBaseByBaseType(ctx context.Context, baseType string) (*interfaces.IndexBaseInfo, error) {
	urlStr := fmt.Sprintf("%s/%s", ibs.appSetting.IndexBaseUrl, baseType)
	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	resCode, res, err := ibs.httpClient.GetNoUnmarshal(ctx, urlStr, nil, headers)
	if err != nil {
		logger.Errorf("failed to get index base by base type, error: %v", err)
		return nil, err
	}

	if resCode != http.StatusOK {
		return nil, fmt.Errorf("response code is %d", resCode)
	}

	var resp []*interfaces.IndexBaseInfo
	err = sonic.Unmarshal(res, &resp)
	if err != nil {
		logger.Errorf("failed to unmarshal get index base by base type, error: %v", err)
		return nil, err
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no index base found for base type %s", baseType)
	}

	return resp[0], nil
}
