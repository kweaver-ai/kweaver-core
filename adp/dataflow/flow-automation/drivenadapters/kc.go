package drivenadapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common/kccode"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

const (
	ArticleSourceHtml      = 0
	ArticleSourceAs        = 1
	ArticleSourceAutoSheet = 2
	ArticleSourceGroup     = 3
)

// Kcmc method interface
type Kcmc interface {
	// GetUserEntity 获取kc用户信息
	GetUserEntity(ctx context.Context, userid, token string) (*userResStruct, error)
	GetArticleByProxyDirID(ctx context.Context, proxyDirID string) (*Article, error)
	IsArticleProxyDocLibSubtype(ctx context.Context, id string) (result bool, err error)
	// 设置文章权限 result 0 成功 1 失败
	SetPerm(ctx context.Context, perm KcPerm, token string) (result float64, err error)
}

type kcmc struct {
	url        string
	urlPrivate string
	httpClient otelHttp.HTTPClient
	conf       rds.ConfDao
}

var (
	kcOnce sync.Once
	kc     Kcmc
)

type userResStruct struct {
	Code int                `json:"code"`
	Msg  string             `json:"msg"`
	Data common.UserInfoMsg `json:"data"`
}

type KcResponse[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

// NewKcmc Kcmc
func NewKcmc() Kcmc {
	kcOnce.Do(func() {
		config := common.NewConfig()
		kc = &kcmc{
			url:        fmt.Sprintf("http://%s:%v", config.Kcmc.Host, config.Kcmc.Port),
			urlPrivate: fmt.Sprintf("http://%s:%v", config.Kcmc.PrivateHost, config.Kcmc.PrivatePort),
			httpClient: NewOtelHTTPClient(),
			conf:       rds.GetConfDao(),
		}
	})
	return kc
}

// GetUserEntity 获取kc用户信息
func (kc *kcmc) GetUserEntity(ctx context.Context, userid, token string) (*userResStruct, error) {
	target := fmt.Sprintf("%s/api/kc-mc/v2/graph-user-info?as_id=%s", kc.url, userid)

	headers := map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
	_, respParam, err := kc.httpClient.Get(ctx, target, headers)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetUserEntity failed: %v, url: %v", err, target)
		return nil, err
	}

	var userRes userResStruct
	respByte, err := json.Marshal(respParam)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.GetUserEntity] Marshal faild, detail: %s", err)
		return nil, err
	}
	err = json.Unmarshal(respByte, &userRes)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.GetUserEntity] Unmarshal faild, detail: %s", err)
		return nil, err
	}

	return &userRes, err
}

type Article struct {
	ArticleID       uint64   `json:"article_id,string"`          // 文章ID
	Thumbnail       string   `json:"thumbnail"`                  // 封面图片地址
	AsID            string   `json:"as_id"`                      // 文章作者asID，创建者的含义
	Title           string   `json:"title"`                      // 标题
	TagIDList       []int    `json:"tag_ids"`                    // 标签ID数组
	TagArr          []string `json:"tags"`                       // 标签数组
	Digest          string   `json:"digest"`                     // 摘要
	Status          int8     `json:"status"`                     // 状态 -10:隐藏 -1:草稿 1:默认 10：已公开,已发布
	DocID           string   `json:"doc_id"`                     // 真实文件的docID 当source=1时即as文件的docid
	ProxyDirID      string   `json:"proxy_dir_id"`               // 知识仓库中代理目录的objectID
	ContentUpdate   int64    `json:"content_update"`             // 内容更新时间戳，已经在业务上定义：只有更新了此文章的正文才叫此文章被更新了，仅其他字段的更新则不认为此文章被更新了。
	ContentSyncTime int64    `json:"content_sync_time"`          // 内容同步到es中的时间
	LikeCount       int32    `json:"like_count"`                 // 点赞数量
	CollectCount    int32    `json:"collect_count"`              // 收藏数量
	CommentCount    int32    `json:"comment_count"`              // 评论数量
	ReplyCount      int32    `json:"reply_count"`                // 回复数量 = 所有评论回复数之和
	ViewCount       int32    `json:"view_count"`                 // 阅读数量
	TopicCount      int32    `json:"topic_count"`                // 关联主题数量  由主题开发组写数据库
	CreateTime      int64    `json:"create_time"`                // 文章原发布时间 仅表示创建此记录的时间戳，由于在发布时可以设置发布时间由insert_time记录
	UpdateTime      int64    `json:"update_time"`                // 更新时间戳, 此字段仅作数据更新标记，不作为文章业务更新的标记，一般不在业务中使用
	FollowCount     int32    `json:"follow_count"`               // 关注数，即有多少个用户关注了此文章
	RecentlyTime    int64    `json:"recently_time"`              // 最近一条评论的创建时间，用于热榜
	FollowTopTime   int64    `json:"follow_top_time"`            // 关注的文章被置顶的时间
	PublisherID     string   `json:"publisher_id"`               // 发布者asID, 只有代发场景才会有，否则和创建者一致，并且后续编辑时不能更新
	InsertTime      int64    `json:"insert_time"`                // 文章原发布时间 发布时可以设置发布时间由此字段记录, 默认=create_time
	ShareCount      int32    `json:"share_count"`                // 分享数
	SpaceID         uint64   `json:"space_id,string"`            // 所属空间id, 默认为0即不属于任何空间
	SpacePath       string   `json:"space_path" `                // 空间路径, 如 /123/456, 操作状态的文章这里保证有路径,但space_dir表中不会有记录
	Source          int      `json:"source"`                     // 对象来源 0:html的wikidoc 1:as文件  2:数据表
	Suffix          string   `json:"suffix"`                     // 当source=1时的后缀
	SpacePathText   string   `json:"space_path_text"`            // 空间路径文本, 如 /空间1/目录1/目录2
	CanArticleWrite bool     `json:"can_article_write"`          // wikidoc是否具有编辑(仅能编辑)权限
	CanArticleRead  *bool    `json:"can_article_read,omitempty"` // wikidoc是否具有预览权限，仅详情接口有此字段
}

func (kc *kcmc) GetArticleByProxyDirID(ctx context.Context, proxyDirID string) (*Article, error) {
	proxyObjectID := utils.GetDocCurID(proxyDirID)
	target := fmt.Sprintf("%s/api/pri-kc-mc/v1/article?proxy_dir_id=%s", kc.urlPrivate, proxyObjectID)

	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}

	_, respParam, err := kc.httpClient.Get(ctx, target, headers)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetArticleByProxyDirID failed: %v, url: %v", err, target)
		return nil, err
	}

	var kcResp KcResponse[Article]
	respByte, _ := json.Marshal(respParam)

	err = json.Unmarshal(respByte, &kcResp)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.GetArticleByProxyDirID] Unmarshal faild, detail: %s", err)
		return nil, err
	}

	if kcResp.Code != kccode.Success {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.GetArticleByProxyDirID] failed, detail: %s", kcResp.Msg)
		return nil, errors.New(kcResp.Msg)
	}

	return &kcResp.Data, nil
}

type KcOem struct {
	CustomDocLibTypeID string `json:"oem:custom_doc_lib_type_id"`
}

func (kc *kcmc) GetKcOem(ctx context.Context) (oem *KcOem, err error) {
	target := fmt.Sprintf("%s/api/kc-mc/v1/oem", kc.url)

	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}

	_, respParam, err := kc.httpClient.Get(ctx, target, headers)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetKcOem failed: %v, url: %v", err, target)
		return
	}

	var kcResp KcResponse[KcOem]
	respByte, _ := json.Marshal(respParam)

	err = json.Unmarshal(respByte, &kcResp)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.GetKcOem] Unmarshal faild, detail: %s", err)
		return nil, err
	}

	if kcResp.Code != kccode.Success {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.GetKcOem] failed, detail: %s", kcResp.Msg)
		return nil, errors.New(kcResp.Msg)
	}

	return &kcResp.Data, nil
}

func (kc *kcmc) IsArticleProxyDocLibSubtype(ctx context.Context, id string) (result bool, err error) {
	subtype, err := kc.conf.Get(ctx, common.ConfKeyArticleProxyDocLibSubtype)
	if err != nil {
		return
	}

	if subtype == "" {

		m, err := NewPersonalConfig().GetModuleByName(ctx, "knowledgecenter")

		if err != nil {
			return false, err
		}

		if m.Name != "KnowledgeCenter" {
			return false, nil
		}

		oem, err := kc.GetKcOem(ctx)
		if err != nil {
			return false, err
		}
		_ = kc.conf.Set(ctx, common.ConfKeyArticleProxyDocLibSubtype, oem.CustomDocLibTypeID)
		subtype = oem.CustomDocLibTypeID
	}

	result = subtype == id

	return
}

type KcPermItem struct {
	EndTime    int64  `json:"end_time"`
	Kind       string `json:"kind"`
	ObjectID   string `json:"object_id"`
	ObjectType string `json:"object_type"`
}

type KcPerm struct {
	Link     string        `json:"link"`
	PowerArr []*KcPermItem `json:"power_arr"`
	SAID     string        `json:"sa_id"`
}

func (kc *kcmc) SetPerm(ctx context.Context, perm KcPerm, token string) (result float64, err error) {
	target := fmt.Sprintf("%s/api/kc-mc/v2/space-tree-perm", kc.url)

	if !strings.HasPrefix(token, "Bearer") {
		token = fmt.Sprintf("Bearer %s", token)
	}

	headers := map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"Authorization": token,
	}

	_, respParam, err := kc.httpClient.Post(ctx, target, headers, perm)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.KcSetPerm] failed: %v, url: %v", err, target)
		return 1, err
	}

	var kcResp KcResponse[string]
	respByte, _ := json.Marshal(respParam)
	err = json.Unmarshal(respByte, &kcResp)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.KcSetPerm] Unmarshal faild, detail: %s", err)
		return 1, err
	}

	if kcResp.Code != kccode.Success {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.KcSetPerm] failed, detail: %s", kcResp.Msg)
		return 1, errors.New(kcResp.Msg)
	}

	return 0, nil
}
