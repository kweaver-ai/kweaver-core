package drivenadapters

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/store"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type StructureExtractor interface {
	FileParse(ctx context.Context, fileUrl, fileName string) (*FileParseResultItem, []*ContentItem, error)
}

type structureExtractor struct {
	baseURL string
	config  *common.StructureExtractor
	client  *http.Client
	client2 HTTPClient2
}

var (
	structureExtractorIns  StructureExtractor
	structureextractorOnce sync.Once
)

func NewStructureExtractor() StructureExtractor {
	structureextractorOnce.Do(func() {
		config := common.NewConfig().StructureExtractor
		structureExtractorIns = &structureExtractor{
			baseURL: fmt.Sprintf("http://%s:%v", config.PrivateHost, config.PrivatePort),
			config:  &config,
			client: &http.Client{
				Transport: otelhttp.NewTransport(&http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}),
				Timeout: 30 * time.Minute, // 整体请求超时（文件解析可能耗时较长）
			},
			client2: NewHTTPClient2(),
		}
	})

	return structureExtractorIns
}

type FileParseResultItem struct {
	Images      map[string]string `json:"images"`
	MdContent   string            `json:"md_content"`
	ContentList string            `json:"content_list"`
}

type ContentItem struct {
	// 通用字段 (所有类型都有)
	Type    string `json:"type"`     // text, image, table, equation, code, list, header, footer, etc.
	PageIdx int    `json:"page_idx"` // 页码,从 0 开始
	Bbox    [4]int `json:"bbox"`     // [x0, y0, x1, y1], 归一化到 0-1000

	// 文本类型字段
	Text      string `json:"text,omitempty"`       // 文本内容
	TextLevel *int   `json:"text_level,omitempty"` // 标题层级,0=正文,1-N=标题

	// 图片类型字段
	ImgPath       string   `json:"img_path,omitempty"`       // 图片路径
	ImageCaption  []string `json:"image_caption,omitempty"`  // 图片描述
	ImageFootnote []string `json:"image_footnote,omitempty"` // 图片脚注

	// 公式类型字段
	TextFormat string `json:"text_format,omitempty"` // "latex"

	// 表格类型字段
	TableCaption  []string `json:"table_caption,omitempty"`  // 表格标题
	TableFootnote []string `json:"table_footnote,omitempty"` // 表格脚注
	TableBody     string   `json:"table_body,omitempty"`     // HTML 格式表格

	// 代码类型字段 (VLM 后端)
	SubType     string   `json:"sub_type,omitempty"`     // "code", "algorithm", "text", "ref_text"
	CodeCaption []string `json:"code_caption,omitempty"` // 代码标题
	CodeBody    string   `json:"code_body,omitempty"`    // 代码内容
	GuessLang   string   `json:"guess_lang,omitempty"`   // 编程语言

	// 列表类型字段 (VLM 后端)
	ListItems []string `json:"list_items,omitempty"` // 列表项
}

type FileParseResult struct {
	Results map[string]*FileParseResultItem `json:"results"`
}

type MineruFile struct {
	Name       string `json:"name"`
	DataID     string `json:"data_id"`
	PageRanges string `json:"page_ranges"`
}

type MineruFileUrlsBatchReq struct {
	IsOCR         bool         `json:"is_ocr"`
	EnableFormula bool         `json:"enable_formula"`
	EnableTable   bool         `json:"enable_table"`
	ModelVersion  string       `json:"model_version"`
	Language      *string      `json:"language,omitempty"`
	IsChem        bool         `json:"is_chem"`
	Files         []MineruFile `json:"files"`
}

type MineruFileUrlsBatchRespData struct {
	BatchID  string              `json:"batch_id"`
	FileURLs []string            `json:"file_urls"`
	TaskIDs  []string            `json:"task_ids"`
	Headers  []map[string]string `json:"headers"`
}

type MineruFileUrlsBatchResp struct {
	Code    int                         `json:"code"`
	Message string                      `json:"msg"`
	TraceID string                      `json:"trace_id"`
	Data    MineruFileUrlsBatchRespData `json:"data"`
}

type MineruTaskFileInfo struct {
	Pages    int `json:"pages"`
	FileSize int `json:"file_size"`
}

type MineruTaskResultData struct {
	TaskID       string             `json:"task_id"`
	State        string             `json:"state"` // "pending", "processing", "done", "failed"
	ErrMsg       string             `json:"err_msg"`
	FullZipURL   string             `json:"full_zip_url"`
	LayoutURL    string             `json:"layout_url"`
	FullMDLink   string             `json:"full_md_link"`
	FileName     string             `json:"file_name"`
	URL          string             `json:"url"`
	Type         string             `json:"type"`
	FileInfo     MineruTaskFileInfo `json:"file_info"`
	ModelVersion string             `json:"model_version"`
	IsChem       bool               `json:"is_chem"`
	ImageReady   bool               `json:"image_ready"`
}

type MineruTaskResultResp struct {
	Code    int                  `json:"code"`
	Msg     string               `json:"msg"`
	TraceID string               `json:"trace_id"`
	Data    MineruTaskResultData `json:"data"`
}

// FileParse 根据配置选择使用内部服务或 MinerU 官方 API
func (s *structureExtractor) FileParse(ctx context.Context, fileUrl, fileName string) (*FileParseResultItem, []*ContentItem, error) {
	if s.config.UseMineru {
		return s.FileParseMineru(ctx, fileUrl, fileName)
	}
	return s.FileParseInternal(ctx, fileUrl, fileName)
}

// FileParseInternal 使用内部部署的 mineru 服务解析文件
func (s *structureExtractor) FileParseInternal(ctx context.Context, fileUrl, fileName string) (*FileParseResultItem, []*ContentItem, error) {
	log := traceLog.WithContext(ctx)

	// 1. 先下载文件（建立连接，获取 response body）
	fileResp, err := s.client.Get(fileUrl)
	if err != nil {
		log.Warnf("FileParse download file err: %s, url: %s", err.Error(), fileUrl)
		return nil, nil, err
	}
	defer fileResp.Body.Close()

	// 2. 使用 io.Pipe 进行流式传输（避免大文件占用过多内存）
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// 先获取 Content-Type（必须在写入之前获取，确保 boundary 正确）
	contentType := writer.FormDataContentType()

	// 用于捕获 goroutine 中的错误
	errChan := make(chan error, 1)

	go func() {
		var writeErr error
		defer func() {
			writer.Close() // 先关闭 writer
			if writeErr != nil {
				pw.CloseWithError(writeErr)
			} else {
				pw.Close()
			}
			errChan <- writeErr
		}()

		// 写入表单字段（用 map+循环简化逻辑）
		fields := map[string]string{
			"output_dir":          s.config.OutputDir,
			"backend":             s.config.Backend,
			"server_url":          s.config.ServerUrl,
			"lang_list":           "ch",
			"return_md":           "true",
			"return_content_list": "true",
			"return_images":       "false",
		}
		for k, v := range fields {
			if writeErr = writer.WriteField(k, v); writeErr != nil {
				return
			}
		}

		// 如果 fileName 是 word 或 excel 文件，在后面拼接 .pdf（只用于表单传递）
		formFileName := fileName
		fileExt := utils.GetFileExtension(fileName)
		if fileExt == ".doc" || fileExt == ".docx" || fileExt == ".xls" || fileExt == ".xlsx" {
			formFileName = fileName + ".pdf"
		}

		// 创建文件表单字段
		part, err := writer.CreateFormFile("files", formFileName)
		if err != nil {
			writeErr = err
			return
		}

		// 流式复制文件内容
		_, writeErr = io.Copy(part, fileResp.Body)
	}()

	// 3. 发送请求（会从 pipe reader 读取数据）
	target := fmt.Sprintf("%s/file_parse", s.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", target, pr)
	if err != nil {
		pr.Close() // 关闭 reader，让 goroutine 退出
		<-errChan  // 等待 goroutine 结束
		log.Warnf("FileParse create request err: %s", err.Error())
		return nil, nil, err
	}

	req.Header.Set("Content-Type", contentType)

	log.Infof("FileParse sending request to %s, Content-Type: %s, fileName: %s", target, contentType, fileName)
	resp, err := s.client.Do(req)

	// 等待写入 goroutine 完成，获取写入错误
	writeErr := <-errChan
	if writeErr != nil {
		log.Warnf("FileParse write goroutine error: %s", writeErr.Error())
		if err == nil {
			err = writeErr
		}
	}

	if err != nil {
		traceLog.WithContext(ctx).Warnf("FileParse err: %s", err.Error())
		return nil, nil, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("FileParse err: %s", err.Error())
		return nil, nil, err
	}

	respCode := resp.StatusCode
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		err = errors.ExHTTPError{
			Body:   string(data),
			Status: respCode,
		}
		return nil, nil, err
	}

	var res FileParseResult
	err = json.Unmarshal(data, &res)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("FileParse err: %s", err.Error())
		return nil, nil, err
	}

	for _, item := range res.Results {
		var contentList []*ContentItem
		if err := json.Unmarshal([]byte(item.ContentList), &contentList); err != nil {
			return item, nil, nil
		}

		return item, contentList, nil
	}

	return nil, nil, nil
}

// FileParseMineru 使用 MinerU 官方 API 解析文件（批量上传方案）
func (s *structureExtractor) FileParseMineru(ctx context.Context, fileUrl, fileName string) (*FileParseResultItem, []*ContentItem, error) {
	log := traceLog.WithContext(ctx)

	// 1. 从 fileUrl 下载文件（流式）
	fileResp, err := s.client.Get(fileUrl)
	if err != nil {
		log.Warnf("FileParseMineru download source file err: %s, url: %s", err.Error(), fileUrl)
		return nil, nil, err
	}
	defer fileResp.Body.Close()

	if fileResp.StatusCode < http.StatusOK || fileResp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, fmt.Errorf("download source file failed: status=%d", fileResp.StatusCode)
	}

	// 2. 获取上传地址（POST 请求）
	uploadInfo, err := s.getMineruUploadInfo(ctx, &MineruFileUrlsBatchReq{
		IsOCR:         false,
		EnableFormula: true,
		EnableTable:   true,
		IsChem:        false,
		Files: []MineruFile{
			{
				Name:       fileName,
				DataID:     store.NextStringID(),
				PageRanges: "",
			},
		},
		ModelVersion: "vlm",
	})
	if err != nil {
		log.Warnf("FileParseMineru get upload url err: %s", err.Error())
		return nil, nil, err
	}

	taskID := uploadInfo.Data.TaskIDs[0]

	// 获取预签名 URL 和对应的 Content-Type
	uploadURL := uploadInfo.Data.FileURLs[0]
	uploadHeaders := map[string]string{}
	if len(uploadInfo.Data.Headers) > 0 {
		uploadHeaders = uploadInfo.Data.Headers[0]
	}

	// 3. 上传文件到预签名地址（使用正确的 Content-Type）
	if err := s.uploadFileToMineru(ctx, uploadURL, uploadHeaders, fileResp.Body); err != nil {
		log.Warnf("FileParseMineru upload file err: %s, task_id: %s", err.Error(), taskID)
		return nil, nil, err
	}

	log.Infof("FileParseMineru file uploaded, task_id: %s, fileName: %s", taskID, fileName)

	// 4. 轮询任务结果
	result, err := s.pollMineruTaskResult(ctx, taskID)
	if err != nil {
		log.Warnf("FileParseMineru poll task result err: %s, task_id: %s", err.Error(), taskID)
		return nil, nil, err
	}

	// 5. 下载并解析 zip 文件
	parseResult, contentList, err := s.downloadAndExtractMineruResult(ctx, result.FullZipURL)
	if err != nil {
		log.Warnf("FileParseMineru download/extract zip err: %s", err.Error())
		return nil, nil, err
	}

	log.Infof("FileParseMineru task completed, task_id: %s, fileName: %s", taskID, fileName)
	return parseResult, contentList, nil
}

// getMineruUploadInfo 获取 MinerU 文件上传信息（POST 请求）
func (s *structureExtractor) getMineruUploadInfo(ctx context.Context, reqBody *MineruFileUrlsBatchReq) (*MineruFileUrlsBatchResp, error) {
	target := fmt.Sprintf("%s/api/v4/file-urls/batch", s.config.MineruBaseURL)
	var resp MineruFileUrlsBatchResp
	respCode, err := s.client2.Post(ctx, target, map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", s.config.MineruToken),
	}, reqBody, &resp)

	if err != nil {
		return nil, err
	}

	if respCode < http.StatusOK || respCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("get upload url failed: status=%d, body=%+v", respCode, resp)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("mineru api error: code=%d, msg=%s", resp.Code, resp.Message)
	}

	if len(resp.Data.TaskIDs) == 0 || len(resp.Data.FileURLs) == 0 {
		return nil, fmt.Errorf("invalid response: missing task_ids or file_urls, body: %+v", resp)
	}

	return &resp, nil
}

// uploadFileToMineru 上传文件到 MinerU 预签名地址
func (s *structureExtractor) uploadFileToMineru(ctx context.Context, uploadURL string, headers map[string]string, fileContent io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, fileContent)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: status=%d, body=%s", resp.StatusCode, string(data))
	}

	return nil
}

func (s *structureExtractor) pollMineruTaskResult(ctx context.Context, taskID string) (*MineruTaskResultData, error) {
	pollInterval := 2 * time.Second
	maxWaitTime := 30 * time.Minute
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if time.Since(startTime) > maxWaitTime {
			return nil, fmt.Errorf("task timeout after %v", maxWaitTime)
		}

		result, err := s.getMineruTaskResult(ctx, taskID)
		if err != nil {
			return nil, err
		}

		switch result.State {
		case "done":
			return result, nil
		case "failed":
			return nil, fmt.Errorf("task failed: %s", result.ErrMsg)
		default:
			time.Sleep(pollInterval)
		}
	}
}

func (s *structureExtractor) getMineruTaskResult(ctx context.Context, taskID string) (*MineruTaskResultData, error) {
	target := fmt.Sprintf("%s/api/v4/extract/task/%s", s.config.MineruBaseURL, taskID)
	var resp MineruTaskResultResp
	respCode, err := s.client2.Get(ctx, target, map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.config.MineruToken),
	}, &resp)

	if err != nil {
		return nil, err
	}

	if respCode < http.StatusOK || respCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("get task result failed: status=%d, body=%+v", respCode, resp)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("mineru api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return &resp.Data, nil
}

// downloadAndExtractMineruResult 下载并解压 zip 文件，提取 content_list 和 md_content
func (s *structureExtractor) downloadAndExtractMineruResult(ctx context.Context, zipURL string) (*FileParseResultItem, []*ContentItem, error) {
	// 下载 zip 文件
	resp, err := s.client.Get(zipURL)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, fmt.Errorf("download zip failed: status=%d", resp.StatusCode)
	}

	// 读取 zip 内容
	zipData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	// 解压 zip
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, nil, err
	}

	var mdContent string
	var contentListStr string
	var contentList []*ContentItem

	for _, file := range zipReader.File {
		// 查找 full.md 文件
		if strings.HasSuffix(file.Name, "full.md") || strings.HasSuffix(file.Name, "/full.md") {
			rc, err := file.Open()
			if err != nil {
				return nil, nil, err
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, nil, err
			}
			mdContent = string(data)
		}

		// 查找 content_list.json 文件（文件名格式：{uuid}_content_list.json）
		if strings.Contains(file.Name, "_content_list.json") {
			rc, err := file.Open()
			if err != nil {
				return nil, nil, err
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, nil, err
			}
			contentListStr = string(data)
			if err := json.Unmarshal(data, &contentList); err != nil {
				// 解析失败，保留原始字符串
				contentList = nil
			}
		}
	}

	if mdContent == "" && contentListStr == "" {
		return nil, nil, fmt.Errorf("no content_list.json or full.md found in zip")
	}

	result := &FileParseResultItem{
		MdContent:   mdContent,
		ContentList: contentListStr,
	}

	return result, contentList, nil
}
