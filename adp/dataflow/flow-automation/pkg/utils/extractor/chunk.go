// Package extractor provides utilities for parsing PDF documents and converting them to structured element formats.
package extractor

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

// SliceType 常量定义
const (
	SliceTypeTitle   = 0 // 标题（不生成独立切片，用于上下文）
	SliceTypeText    = 1 // 文本
	SliceTypeTable   = 2 // 表格（原子对象）
	SliceTypeImage   = 3 // 图片（原子对象）
	SliceTypeFormula = 4 // 公式（原子对象）
	SliceTypeCode    = 5 // 代码（原子对象）
	SliceTypeOther   = 6 // 其他
)

// CustomChunk 返回的块结构
// 注意：Chunk 和 Element 不是一一对应的关系：
//   - 一个 Chunk 可能对应多个 Element（合并场景）
//   - 多个 Chunk 可能对应同一个 Element（切割场景）
//
// ElementID: 第一个关联的 Element ID，用于直接建立 Chunk 与 Element 的关系
// ElementIDs: 关联的 Element ID 列表，用于建立 Chunk 与 Element 的完整关系
//   - 对于单个 Element 的 Chunk（标题、表格、图片等），ElementID 和 ElementIDs[0] 相同
//   - 对于合并多个 Element 的文本 Chunk，ElementID 是第一个 ElementID，ElementIDs 包含所有相关的 ElementID
//
// 父子关系通过 Element 层获取：根据 ElementID 或 ElementIDs 找到对应的 Element，通过 Element.parent_element_id 获取
type CustomChunk struct {
	DocName         string    `json:"doc_name"`
	DocMD5          string    `json:"doc_md5"`
	SliceMD5        string    `json:"slice_md5"`
	ID              string    `json:"id"`
	DeduplicationID string    `json:"deduplication_id"` // 基于内容生成的稳定ID，用于去重
	DocumentID      string    `json:"document_id"`
	SliceType       int       `json:"slice_type"`
	Pages           []int     `json:"pages"`
	SegmentID       int       `json:"segment_id"`            // 对应的第一个 ContentItem/Element 的索引，用于关联父子关系
	ElementID       string    `json:"element_id,omitempty"`  // 第一个关联的 Element ID，用于直接建立关系
	ElementIDs      []string  `json:"element_ids,omitempty"` // 关联的 Element ID 列表，用于建立完整关系
	Location        [][4]int  `json:"location"`
	SliceContent    string    `json:"slice_content"`
	Embedding       []float64 `json:"text_vector"`
	ImgPath         *string   `json:"img_path,omitempty"`
	ImageVector     []float64 `json:"image_vector,omitempty"`
}

// ProcessingItem 处理过程中的中间结构
type ProcessingItem struct {
	Type          string
	PageIdx       interface{} // 可以是 int 或 []int
	Bbox          interface{} // 可以是 [4]int 或 [][4]int
	Text          string
	TextLevel     *int
	ImgPath       string
	ImageCaption  []string
	ImageFootnote []string
	TextFormat    string
	TableCaption  []string
	TableFootnote []string
	TableBody     string
	SubType       string
	CodeCaption   []string
	CodeBody      string
	GuessLang     string
	ListItems     []string

	OriginalIndex         int
	MergedOriginalIndices []int // 用于跟踪合并的原始索引
}

// GenerateMD5 生成文本或字节的 MD5 哈希
func GenerateMD5(data interface{}) string {
	var hash [16]byte
	switch v := data.(type) {
	case string:
		hash = md5.Sum([]byte(v))
	case []byte:
		hash = md5.Sum(v)
	default:
		// 转换为字符串处理
		hash = md5.Sum([]byte(fmt.Sprintf("%v", v)))
	}
	return hex.EncodeToString(hash[:])
}

// GenerateDeduplicationID 生成基于内容的稳定去重ID
// 对于chunk: 使用 docMD5 + sliceMD5 + chunkIndex
// 对于element: 使用 docMD5 + elementIndex + contentHash + elementType + pageIdx
func GenerateDeduplicationID(docMD5 string, parts ...string) string {
	combined := docMD5
	for _, part := range parts {
		combined += "_" + part
	}
	return GenerateMD5(combined)
}

// DeduplicateChunks 基于 deduplication_id 对 chunks 进行去重
// 返回去重后的 chunks 和重复的 deduplication_id 列表
func DeduplicateChunks(chunks []*CustomChunk) ([]*CustomChunk, []string) {
	seen := make(map[string]bool)
	deduplicated := make([]*CustomChunk, 0)
	duplicates := make([]string, 0)

	for _, chunk := range chunks {
		if chunk.DeduplicationID == "" {
			// 如果没有去重ID，保留该chunk
			deduplicated = append(deduplicated, chunk)
			continue
		}

		if seen[chunk.DeduplicationID] {
			// 发现重复
			duplicates = append(duplicates, chunk.DeduplicationID)
			continue
		}

		seen[chunk.DeduplicationID] = true
		deduplicated = append(deduplicated, chunk)
	}

	return deduplicated, duplicates
}

// determineSliceType 根据项目类型确定切片类型
func determineSliceType(item *ProcessingItem) int {
	// 1. 优先处理文本级别为标题的情况
	if item.TextLevel != nil && (*item.TextLevel >= 1 && *item.TextLevel <= 3) {
		return 0 // 标题标识符
	}

	// 2. 类型映射
	typeMap := map[string]int{
		"text":     1, // 文本
		"list":     1, // 列表（作为文本处理）
		"table":    2, // 表格
		"image":    3, // 图片
		"formula":  4, // 公式
		"equation": 4, // 公式
		"code":     5, // 代码
	}

	// 3. 匹配映射表
	if sliceType, exists := typeMap[strings.ToLower(item.Type)]; exists {
		return sliceType
	}

	// 4. 兜底处理
	if item.Text != "" {
		return 1
	}
	return 6 // 兜底处理
}

// convertToProcessingItems 将 ContentItem 转换为 ProcessingItem
func convertToProcessingItems(contentList []*drivenadapters.ContentItem) []*ProcessingItem {
	processed := make([]*ProcessingItem, len(contentList))
	for i, item := range contentList {
		processed[i] = &ProcessingItem{
			Type:          item.Type,
			PageIdx:       item.PageIdx,
			Bbox:          item.Bbox,
			Text:          item.Text,
			TextLevel:     item.TextLevel,
			ImgPath:       item.ImgPath,
			ImageCaption:  item.ImageCaption,
			ImageFootnote: item.ImageFootnote,
			TextFormat:    item.TextFormat,
			TableCaption:  item.TableCaption,
			TableFootnote: item.TableFootnote,
			TableBody:     item.TableBody,
			SubType:       item.SubType,
			CodeCaption:   item.CodeCaption,
			CodeBody:      item.CodeBody,
			GuessLang:     item.GuessLang,
			ListItems:     item.ListItems,
			OriginalIndex: i,
		}
	}
	return processed
}

// ensureBasicFields 确保基本字段存在并分配原始索引
func ensureBasicFields(processedList []*ProcessingItem) {
	for i, item := range processedList {
		item.OriginalIndex = i

		// 确保文本字段存在
		if item.Text == "" {
			item.Text = ""
		}

		// 推断类型
		if item.Type == "" {
			if item.ImgPath != "" && item.TableBody == "" {
				item.Type = "image"
			} else if item.TableBody != "" {
				item.Type = "table"
			} else if item.TextFormat == "latex" {
				item.Type = "equation"
			} else {
				item.Type = "text"
			}
		}

		// 确保页面索引存在
		if item.PageIdx == nil {
			item.PageIdx = -1
		}

		// 确保边界框存在
		if item.Bbox == nil {
			item.Bbox = [][4]int{}
		}
	}
}

// preprocessItems 预处理图片和表格项目以合并文本
func preprocessItems(processedList []*ProcessingItem) {
	for _, item := range processedList {
		textParts := []string{}
		currentText := strings.TrimSpace(item.Text)

		// code 类型优先使用 CodeBody
		if item.Type == "code" && item.CodeBody != "" {
			currentText = strings.TrimSpace(item.CodeBody)
		}

		if currentText != "" {
			textParts = append(textParts, currentText)
		}

		switch item.Type {
		case "image":
			caption := strings.Join(item.ImageCaption, "\n")
			footnote := strings.Join(item.ImageFootnote, "\n")
			if caption != "" {
				textParts = append(textParts, caption)
			}
			if footnote != "" {
				textParts = append(textParts, footnote)
			}
		case "table":
			caption := strings.Join(item.TableCaption, "\n")
			body := strings.TrimSpace(item.TableBody)
			footnote := strings.Join(item.TableFootnote, "\n")
			if caption != "" {
				textParts = append(textParts, caption)
			}
			if body != "" {
				textParts = append(textParts, body)
			}
			if footnote != "" {
				textParts = append(textParts, footnote)
			}
		case "code":
			caption := strings.Join(item.CodeCaption, "\n")
			if caption != "" {
				textParts = append(textParts, caption)
			}
		case "list":
			// 处理列表项，将 list_items 合并为文本
			if len(item.ListItems) > 0 {
				listContent := strings.Join(item.ListItems, "\n")
				if listContent != "" {
					textParts = append(textParts, listContent)
				}
			}
		}

		// 去重
		uniqueParts := []string{}
		seen := make(map[string]bool)
		for _, part := range textParts {
			strippedPart := strings.TrimSpace(part)
			if strippedPart != "" && !seen[strippedPart] {
				uniqueParts = append(uniqueParts, strippedPart)
				seen[strippedPart] = true
			}
		}
		item.Text = strings.Join(uniqueParts, "\n")
	}
}

// findImageRelatedOCRElements 查找与图片位置相关的 OCR 文本元素
// 返回应该合并到图片 chunk 中的元素索引列表
func findImageRelatedOCRElements(imageItem *ProcessingItem, processedList []*ProcessingItem, startIdx int, endIdx int) []int {
	var relatedIndices []int

	// 获取图片的 bbox 和 pageIdx
	var imageBbox [4]int
	var imagePageIdx int

	switch b := imageItem.Bbox.(type) {
	case [4]int:
		imageBbox = b
	case [][4]int:
		if len(b) > 0 {
			imageBbox = b[0]
		} else {
			return relatedIndices
		}
	default:
		return relatedIndices
	}

	switch p := imageItem.PageIdx.(type) {
	case int:
		imagePageIdx = p
	case []int:
		if len(p) > 0 {
			imagePageIdx = p[0]
		} else {
			return relatedIndices
		}
	default:
		return relatedIndices
	}

	// 图片的底部 y 坐标
	imageBottomY := imageBbox[3]

	// 查找图片下方一定范围内的文本元素（OCR 文本通常是图片下方的说明文字）
	// 范围：图片下方 200 像素内（归一化坐标，0-1000）
	maxDistance := 200

	// 从图片后面的元素开始查找
	for i := startIdx + 1; i < endIdx && i < len(processedList); i++ {
		item := processedList[i]

		// 跳过图片本身
		if item.Type == "image" {
			continue
		}

		// 只处理文本类型的元素
		if item.Type != "text" && item.Type != "title" {
			break // 遇到非文本元素（如表格、其他图片）则停止
		}

		// 检查是否在同一页面
		var itemPageIdx int
		switch p := item.PageIdx.(type) {
		case int:
			itemPageIdx = p
		case []int:
			if len(p) > 0 {
				itemPageIdx = p[0]
			} else {
				continue
			}
		default:
			continue
		}

		if itemPageIdx != imagePageIdx {
			break // 不在同一页面，停止查找
		}

		// 获取元素的 bbox
		var itemBbox [4]int
		switch b := item.Bbox.(type) {
		case [4]int:
			itemBbox = b
		case [][4]int:
			if len(b) > 0 {
				itemBbox = b[0]
			} else {
				continue
			}
		default:
			continue
		}

		// 检查元素是否在图片下方
		itemTopY := itemBbox[1]
		if itemTopY < imageBottomY {
			continue // 元素在图片上方，跳过
		}

		// 检查距离是否在范围内
		distance := itemTopY - imageBottomY
		if distance > maxDistance {
			break // 距离太远，停止查找
		}

		// 检查水平位置是否重叠或接近（允许一定的水平偏移）
		// 图片的左右边界
		imageLeftX := imageBbox[0]
		imageRightX := imageBbox[2]
		// 元素的左右边界
		itemLeftX := itemBbox[0]
		itemRightX := itemBbox[2]

		// 允许的水平偏移（归一化坐标）
		horizontalTolerance := 100

		// 检查是否有水平重叠或接近
		hasHorizontalOverlap := (itemLeftX <= imageRightX+horizontalTolerance) && (itemRightX >= imageLeftX-horizontalTolerance)

		if hasHorizontalOverlap {
			// 检查是否是标题（标题通常不应该合并到图片中）
			if item.TextLevel != nil && *item.TextLevel > 0 {
				// 如果是标题，停止查找
				break
			}

			// 找到相关的 OCR 元素
			relatedIndices = append(relatedIndices, i)
		}
	}

	return relatedIndices
}

// identifySplitPoints 识别分割点
func identifySplitPoints(processedList []*ProcessingItem) (map[int]bool, []int) {
	splitIndices := make(map[int]bool)
	var level1HeadingIndices []int

	if len(processedList) > 0 {
		splitIndices[0] = true // 文档开始
	}

	// 页面断点分割
	lastProcessedPage := -2
	for i, item := range processedList {
		var pageIdx int
		switch p := item.PageIdx.(type) {
		case int:
			pageIdx = p
		case []int:
			if len(p) > 0 {
				pageIdx = p[0]
			} else {
				pageIdx = -1
			}
		default:
			pageIdx = -1
		}

		if pageIdx != -1 && pageIdx != lastProcessedPage && i > 0 {
			splitIndices[i] = true
		}
		if pageIdx != -1 {
			lastProcessedPage = pageIdx
		}

		// 一级标题分割点
		if item.TextLevel != nil && *item.TextLevel == 1 {
			splitIndices[i] = true
			level1HeadingIndices = append(level1HeadingIndices, i)
		}
	}

	return splitIndices, level1HeadingIndices
}

// mergeAdjacentHeadings 合并相邻的一级标题
func mergeAdjacentHeadings(processedList []*ProcessingItem, level1HeadingIndices []int, splitIndices map[int]bool) (map[int]bool, map[int]*ProcessingItem) {
	mergedIndicesToRemove := make(map[int]bool)
	mergedHeadingMap := make(map[int]*ProcessingItem)

	if len(level1HeadingIndices) == 0 {
		return mergedIndicesToRemove, mergedHeadingMap
	}

	sort.Ints(level1HeadingIndices)
	currentMergeGroup := []int{}
	lastHeadingIdx := -2

	processMergeGroup := func(group []int) {
		if len(group) == 0 {
			return
		}

		if len(group) > 1 {
			firstIdx := group[0]
			mergedItem := &ProcessingItem{
				Type:      "text",
				TextLevel: intPtr(1),
			}

			mergedTexts := []string{}
			mergedPagesSet := make(map[int]bool)
			mergedBboxes := [][4]int{}
			originalIndices := []int{}

			for _, idx := range group {
				item := processedList[idx]
				originalIndices = append(originalIndices, item.OriginalIndex)

				textToAppend := strings.TrimSpace(item.Text)
				if textToAppend != "" {
					mergedTexts = append(mergedTexts, textToAppend)
				}

				// 处理页面
				switch p := item.PageIdx.(type) {
				case int:
					if p != -1 {
						mergedPagesSet[p] = true
					}
				case []int:
					for _, page := range p {
						if page != -1 {
							mergedPagesSet[page] = true
						}
					}
				}

				// 处理边界框
				switch b := item.Bbox.(type) {
				case [4]int:
					mergedBboxes = append(mergedBboxes, b)
				case [][4]int:
					mergedBboxes = append(mergedBboxes, b...)
				}

				if idx != firstIdx {
					mergedIndicesToRemove[idx] = true
				}
			}

			mergedItem.Text = strings.Join(mergedTexts, "\n")

			// 转换页面集合为排序后的切片
			mergedPages := make([]int, 0, len(mergedPagesSet))
			for page := range mergedPagesSet {
				mergedPages = append(mergedPages, page)
			}
			sort.Ints(mergedPages)
			mergedItem.PageIdx = mergedPages

			mergedItem.Bbox = mergedBboxes
			mergedItem.OriginalIndex = originalIndices[0]
			mergedItem.MergedOriginalIndices = originalIndices

			mergedHeadingMap[firstIdx] = mergedItem

			// 移除冗余分割点
			for _, idx := range group[1:] {
				delete(splitIndices, idx)
			}
		} else if len(group) == 1 {
			// 确保单个标题的格式一致性
			idx := group[0]
			item := processedList[idx]

			// 确保页面索引是切片格式
			switch p := item.PageIdx.(type) {
			case int:
				if p != -1 {
					item.PageIdx = []int{p}
				} else {
					item.PageIdx = []int{}
				}
			case []int:
				// 已经是切片，不需要处理
			default:
				item.PageIdx = []int{}
			}

			// 确保边界框是切片格式
			switch b := item.Bbox.(type) {
			case [4]int:
				item.Bbox = [][4]int{b}
			case [][4]int:
				// 已经是切片，不需要处理
			default:
				item.Bbox = [][4]int{}
			}
		}
	}

	for _, idx := range level1HeadingIndices {
		isAdjacent := (idx == lastHeadingIdx+1)
		if isAdjacent {
			currentMergeGroup = append(currentMergeGroup, idx)
		} else {
			processMergeGroup(currentMergeGroup)
			currentMergeGroup = []int{idx}
		}
		lastHeadingIdx = idx
	}
	processMergeGroup(currentMergeGroup)

	return mergedIndicesToRemove, mergedHeadingMap
}

// UploadImageToOSS 上传图片到 OSS 并返回下载链接
func UploadImageToOSS(ctx context.Context, imgPath string, docName string) (string, error) {
	log := traceLog.WithContext(ctx)

	// img_path 只存储路径部分，需要拼接完整的下载 URL
	// 例如：img_path = "file/doc_name/images/76d2c9e76493c89b1b82971e7821d7b6e0b957a8e78928543b3a730a9646a8e2.jpg"
	// 需要拼接成：http://192.168.173.19:8090/file/doc_name/images/76d2c9e76493c89b1b82971e7821d7b6e0b957a8e78928543b3a730a9646a8e2.jpg

	// 从 img_path 中提取文件名（包含哈希和扩展名）
	imageFileName := filepath.Base(imgPath)
	if imageFileName == "" || imageFileName == "." || imageFileName == "/" {
		log.Warnf("[uploadImageToOSS] invalid image file name: %s", imgPath)
		return "", fmt.Errorf("invalid image file name: %s", imgPath)
	}

	// 去掉 docName 的后缀名
	docNameWithoutExt := strings.TrimSuffix(docName, filepath.Ext(docName))

	// 从配置中获取文件服务的host和port
	config := common.NewConfig()
	serverHost := config.StructureExtractor.FileHost
	serverPort := config.StructureExtractor.FilePort

	// 如果配置为空，使用默认值
	if serverHost == "" {
		serverHost = "192.168.173.19"
	}
	if serverPort == "" {
		serverPort = "8090"
	}

	// 拼接完整的下载 URL：http://{host}:{port}/file/{doc_name}/images/{filename}
	// 用于从源服务器下载图片
	downloadURL := fmt.Sprintf("http://%s:%s/file/%s/images/%s", serverHost, serverPort, docNameWithoutExt, imageFileName)

	log.Infof("[uploadImageToOSS] downloading image from URL: %s", downloadURL)

	// 创建 HTTP 客户端下载图片
	client := drivenadapters.NewRawHTTPClient()
	client.Timeout = 60 * time.Second

	resp, err := client.Get(downloadURL)
	if err != nil {
		log.Warnf("[uploadImageToOSS] failed to download image from URL: %s, error: %v", downloadURL, err)
		return "", fmt.Errorf("failed to download image from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Warnf("[uploadImageToOSS] failed to download image, status: %d, body: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("failed to download image, status: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("[uploadImageToOSS] failed to read response body: %s, error: %v", downloadURL, err)
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	imageData := bytes.NewReader(body)
	fileSize := int64(len(body))

	// 计算图片内容的MD5，用于生成稳定的OSS key，避免重复上传相同图片
	imageMD5 := GenerateMD5(body) // 使用图片内容的MD5作为key的一部分

	// 获取文件扩展名
	fileExt := filepath.Ext(imageFileName)
	if fileExt == "" {
		fileExt = ".jpg" // 默认扩展名
	}

	// 生成 OSS key：file/{doc_name}/images/{md5}.{ext}
	// 使用MD5而不是时间戳，相同图片会使用相同的key，避免重复上传
	ossKey := fmt.Sprintf("file/%s/images/%s%s", docNameWithoutExt, imageMD5, fileExt)

	// 获取 OSS 网关实例
	ossGateway := drivenadapters.NewOssGateWay()

	// 获取可用的 OSS ID
	ossID, err := ossGateway.GetAvaildOSS(ctx)
	if err != nil {
		log.Warnf("[uploadImageToOSS] failed to get available OSS: %v", err)
		return "", fmt.Errorf("failed to get available OSS: %w", err)
	}

	// 检查OSS中是否已存在该文件（通过GetObjectMeta检查）
	existingSize, err := ossGateway.GetObjectMeta(ctx, ossID, ossKey, true)
	if err == nil && existingSize > 0 {
		// 文件已存在，直接返回下载URL，避免重复上传
		log.Infof("[uploadImageToOSS] image already exists in OSS, skipping upload: %s", ossKey)
		expires := time.Now().AddDate(1, 0, 0).Unix() // 1年后的Unix时间戳
		resultURL, err := ossGateway.GetDownloadURL(ctx, ossID, ossKey, expires, false)
		if err != nil {
			log.Warnf("[uploadImageToOSS] failed to get download URL for existing file: %s, error: %v", ossKey, err)
			return "", fmt.Errorf("failed to get download URL: %w", err)
		}
		log.Infof("[uploadImageToOSS] reused existing image: %s -> %s", imgPath, resultURL)
		return resultURL, nil
	}

	// 文件不存在，需要上传
	// 重新创建Reader（因为之前已经读取过了）
	imageData = bytes.NewReader(body)
	err = ossGateway.UploadFile(ctx, ossID, ossKey, true, imageData, fileSize)
	if err != nil {
		log.Warnf("[uploadImageToOSS] failed to upload image to OSS: %s, error: %v", imgPath, err)
		return "", fmt.Errorf("failed to upload image to OSS: %w", err)
	}

	// 使用 OSS 方法生成下载链接（设置过期时间为 1 年后）
	expires := time.Now().AddDate(1, 0, 0).Unix() // 1年后的Unix时间戳
	resultURL, err := ossGateway.GetDownloadURL(ctx, ossID, ossKey, expires, false)
	if err != nil {
		log.Warnf("[uploadImageToOSS] failed to get download URL: %s, error: %v", ossKey, err)
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}

	log.Infof("[uploadImageToOSS] successfully uploaded image: %s -> %s", imgPath, resultURL)
	return resultURL, nil
}

// UploadOriginalImageToOSS 从原始文件URL下载图片并上传到OSS
// 用于处理输入文件本身就是图片的情况
func UploadOriginalImageToOSS(ctx context.Context, originalFileURL string, docName string) (string, error) {
	log := traceLog.WithContext(ctx)

	log.Infof("[uploadOriginalImageToOSS] downloading image from URL: %s", originalFileURL)

	// 创建 HTTP 客户端下载图片
	client := drivenadapters.NewRawHTTPClient()
	client.Timeout = 60 * time.Second

	resp, err := client.Get(originalFileURL)
	if err != nil {
		log.Warnf("[uploadOriginalImageToOSS] failed to download image from URL: %s, error: %v", originalFileURL, err)
		return "", fmt.Errorf("failed to download image from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Warnf("[uploadOriginalImageToOSS] failed to download image, status: %d, body: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("failed to download image, status: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("[uploadOriginalImageToOSS] failed to read response body: %s, error: %v", originalFileURL, err)
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	imageData := bytes.NewReader(body)
	fileSize := int64(len(body))

	// 计算图片内容的MD5，用于生成稳定的OSS key，避免重复上传相同图片
	imageMD5 := GenerateMD5(body)

	// 从docName获取文件扩展名，如果没有则使用默认值
	fileExt := filepath.Ext(docName)
	if fileExt == "" {
		fileExt = ".jpg" // 默认扩展名
	}

	// 去掉 docName 的后缀名
	docNameWithoutExt := strings.TrimSuffix(docName, fileExt)

	// 生成 OSS key：file/{doc_name}/images/{md5}.{ext}
	// 使用MD5而不是时间戳，相同图片会使用相同的key，避免重复上传
	ossKey := fmt.Sprintf("file/%s/images/%s%s", docNameWithoutExt, imageMD5, fileExt)

	// 获取 OSS 网关实例
	ossGateway := drivenadapters.NewOssGateWay()

	// 获取可用的 OSS ID
	ossID, err := ossGateway.GetAvaildOSS(ctx)
	if err != nil {
		log.Warnf("[uploadOriginalImageToOSS] failed to get available OSS: %v", err)
		return "", fmt.Errorf("failed to get available OSS: %w", err)
	}

	// 检查OSS中是否已存在该文件（通过GetObjectMeta检查）
	existingSize, err := ossGateway.GetObjectMeta(ctx, ossID, ossKey, true)
	if err == nil && existingSize > 0 {
		// 文件已存在，直接返回下载URL，避免重复上传
		log.Infof("[uploadOriginalImageToOSS] image already exists in OSS, skipping upload: %s", ossKey)
		expires := time.Now().AddDate(1, 0, 0).Unix() // 1年后的Unix时间戳
		resultURL, err := ossGateway.GetDownloadURL(ctx, ossID, ossKey, expires, false)
		if err != nil {
			log.Warnf("[uploadOriginalImageToOSS] failed to get download URL for existing file: %s, error: %v", ossKey, err)
			return "", fmt.Errorf("failed to get download URL: %w", err)
		}
		log.Infof("[uploadOriginalImageToOSS] reused existing image: %s -> %s", originalFileURL, resultURL)
		return resultURL, nil
	}

	// 文件不存在，需要上传
	// 重新创建Reader（因为之前已经读取过了）
	imageData = bytes.NewReader(body)
	err = ossGateway.UploadFile(ctx, ossID, ossKey, true, imageData, fileSize)
	if err != nil {
		log.Warnf("[uploadOriginalImageToOSS] failed to upload image to OSS: %s, error: %v", originalFileURL, err)
		return "", fmt.Errorf("failed to upload image to OSS: %w", err)
	}

	// 使用 OSS 方法生成下载链接（设置过期时间为 1 年后）
	expires := time.Now().AddDate(1, 0, 0).Unix() // 1年后的Unix时间戳
	resultURL, err := ossGateway.GetDownloadURL(ctx, ossID, ossKey, expires, false)
	if err != nil {
		log.Warnf("[uploadOriginalImageToOSS] failed to get download URL: %s, error: %v", ossKey, err)
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}

	log.Infof("[uploadOriginalImageToOSS] successfully uploaded image: %s -> %s", originalFileURL, resultURL)
	return resultURL, nil
}

// createChunks 创建自定义块
func createChunks(ctx context.Context, processedList []*ProcessingItem, splitIndices map[int]bool, mergedIndicesToRemove map[int]bool, mergedHeadingMap map[int]*ProcessingItem, docName, docMD5 string) []*CustomChunk {
	// 获取最终分割点
	finalSplitPoints := make([]int, 0, len(splitIndices))
	for idx := range splitIndices {
		if idx < len(processedList) {
			finalSplitPoints = append(finalSplitPoints, idx)
		}
	}
	sort.Ints(finalSplitPoints)
	customChunks := []*CustomChunk{}

	for i, startIndex := range finalSplitPoints {
		endIndex := len(processedList)
		if i+1 < len(finalSplitPoints) {
			endIndex = finalSplitPoints[i+1]
		}

		chunkContentParts := []string{}
		chunkPagesSet := make(map[int]bool)
		chunkLocations := [][4]int{}
		var firstItemForChunkMetadata *ProcessingItem
		firstItemSegmentID := -1
		var imageImgPath *string

		// 用于跟踪已处理的索引，避免重复处理
		processedIndices := make(map[int]bool)

		currentIndex := startIndex
		for currentIndex < endIndex {
			if mergedIndicesToRemove[currentIndex] {
				currentIndex++
				continue
			}

			if processedIndices[currentIndex] {
				currentIndex++
				continue
			}

			var itemToProcess *ProcessingItem
			if currentIndex == startIndex {
				if mergedItem, exists := mergedHeadingMap[startIndex]; exists {
					itemToProcess = mergedItem
				} else {
					itemToProcess = processedList[currentIndex]
				}
			} else {
				itemToProcess = processedList[currentIndex]
			}

			// 确保格式一致性
			ensureConsistentFormat(itemToProcess)

			if firstItemForChunkMetadata == nil {
				firstItemForChunkMetadata = itemToProcess
				firstItemSegmentID = itemToProcess.OriginalIndex
			}

			// 如果是图片类型，查找相关的 OCR 元素并合并
			if itemToProcess.Type == "image" && itemToProcess.ImgPath != "" {
				// 上传图片到 OSS 并获取下载链接
				uploadedURL, err := UploadImageToOSS(ctx, itemToProcess.ImgPath, docName)
				if err != nil {
					// 如果上传失败，记录警告但继续使用原始路径
					traceLog.WithContext(ctx).Warnf("[createChunks] failed to upload image: %s, error: %v", itemToProcess.ImgPath, err)
					// 即使上传失败，也尝试拼接原始路径的完整 URL
					docNameWithoutExt := strings.TrimSuffix(docName, filepath.Ext(docName))
					// 从配置中获取文件服务的host和port
					config := common.NewConfig()
					serverHost := config.StructureExtractor.FileHost
					serverPort := config.StructureExtractor.FilePort

					// 如果配置为空，使用默认值
					if serverHost == "" {
						serverHost = "192.168.173.19"
					}
					if serverPort == "" {
						serverPort = "8090"
					}
					imageFileName := filepath.Base(itemToProcess.ImgPath)
					fallbackURL := fmt.Sprintf("http://%s:%s/file/%s/images/%s", serverHost, serverPort, docNameWithoutExt, imageFileName)
					imageImgPath = &fallbackURL
				} else {
					imageImgPath = &uploadedURL
				}

				// 查找相关的 OCR 元素
				relatedIndices := findImageRelatedOCRElements(itemToProcess, processedList, currentIndex, endIndex)

				// 处理图片本身
				text := strings.TrimSpace(itemToProcess.Text)
				if itemToProcess.Type == "code" && itemToProcess.CodeBody != "" {
					text = strings.TrimSpace(itemToProcess.CodeBody)
				}
				if text != "" {
					chunkContentParts = append(chunkContentParts, text)
				}

				// 处理页面
				switch p := itemToProcess.PageIdx.(type) {
				case []int:
					for _, page := range p {
						if page != -1 {
							chunkPagesSet[page] = true
						}
					}
				case int:
					if p != -1 {
						chunkPagesSet[p] = true
					}
				}

				// 处理位置
				switch b := itemToProcess.Bbox.(type) {
				case [][4]int:
					chunkLocations = append(chunkLocations, b...)
				case [4]int:
					chunkLocations = append(chunkLocations, b)
				}

				processedIndices[currentIndex] = true

				// 合并相关的 OCR 元素
				for _, relatedIdx := range relatedIndices {
					if relatedIdx >= endIndex || processedIndices[relatedIdx] {
						continue
					}

					relatedItem := processedList[relatedIdx]
					ensureConsistentFormat(relatedItem)

					// 添加 OCR 文本
					ocrText := strings.TrimSpace(relatedItem.Text)
					if ocrText != "" {
						chunkContentParts = append(chunkContentParts, ocrText)
					}

					// 处理页面
					switch p := relatedItem.PageIdx.(type) {
					case []int:
						for _, page := range p {
							if page != -1 {
								chunkPagesSet[page] = true
							}
						}
					case int:
						if p != -1 {
							chunkPagesSet[p] = true
						}
					}

					// 处理位置
					switch b := relatedItem.Bbox.(type) {
					case [][4]int:
						chunkLocations = append(chunkLocations, b...)
					case [4]int:
						chunkLocations = append(chunkLocations, b)
					}

					processedIndices[relatedIdx] = true
				}
			} else {
				// 非图片元素，正常处理
				text := strings.TrimSpace(itemToProcess.Text)
				if itemToProcess.Type == "code" && itemToProcess.CodeBody != "" {
					text = strings.TrimSpace(itemToProcess.CodeBody)
				}
				if text != "" {
					chunkContentParts = append(chunkContentParts, text)
				}

				// 处理页面
				switch p := itemToProcess.PageIdx.(type) {
				case []int:
					for _, page := range p {
						if page != -1 {
							chunkPagesSet[page] = true
						}
					}
				case int:
					if p != -1 {
						chunkPagesSet[p] = true
					}
				}

				// 处理位置
				switch b := itemToProcess.Bbox.(type) {
				case [][4]int:
					chunkLocations = append(chunkLocations, b...)
				case [4]int:
					chunkLocations = append(chunkLocations, b)
				}

				processedIndices[currentIndex] = true
			}

			currentIndex++
		}

		if firstItemForChunkMetadata != nil {
			sliceContent := strings.Join(chunkContentParts, "\n")

			chunkID := uuid.New().String()
			sliceMD5 := GenerateMD5(sliceContent)
			sliceType := determineSliceType(firstItemForChunkMetadata)

			// 生成稳定的去重ID（基于docMD5和sliceMD5）
			deduplicationID := GenerateDeduplicationID(docMD5, sliceMD5, fmt.Sprintf("%d", i))

			// 转换页面集合为排序后的切片
			finalPages := make([]int, 0, len(chunkPagesSet))
			for page := range chunkPagesSet {
				finalPages = append(finalPages, page)
			}
			sort.Ints(finalPages)

			customChunk := &CustomChunk{
				DocName:         docName,
				SliceMD5:        sliceMD5,
				ID:              chunkID,
				DeduplicationID: deduplicationID,
				SliceType:       sliceType,
				Pages:           finalPages,
				SegmentID:       firstItemSegmentID,
				Location:        chunkLocations,
				SliceContent:    sliceContent,
				ImgPath:         imageImgPath,
			}
			customChunks = append(customChunks, customChunk)
		}
	}

	// 添加文档MD5（父子关系将在后续基于 content_list 设置）
	for i := range customChunks {
		customChunks[i].DocMD5 = docMD5
	}

	return customChunks
}

// ensureConsistentFormat 确保处理项格式一致性
func ensureConsistentFormat(item *ProcessingItem) {
	// 确保页面索引是切片格式
	switch p := item.PageIdx.(type) {
	case int:
		if p != -1 {
			item.PageIdx = []int{p}
		} else {
			item.PageIdx = []int{}
		}
	case []int:
		// 已经是切片，不需要处理
	default:
		item.PageIdx = []int{}
	}

	// 确保边界框是切片格式
	switch b := item.Bbox.(type) {
	case [4]int:
		item.Bbox = [][4]int{b}
	case [][4]int:
		// 已经是切片，不需要处理
	default:
		item.Bbox = [][4]int{}
	}
}

// intPtr 返回整数的指针
func intPtr(i int) *int {
	return &i
}

// isNaturalParagraphEnd 判断文本是否以自然段落结尾
// 自然段落结尾：句号、问号、感叹号、冒号后的换行，或以换行符结尾
func isNaturalParagraphEnd(text string) bool {
	if text == "" {
		return false
	}
	trimmed := strings.TrimRight(text, " \t")
	if trimmed == "" {
		return false
	}
	lastChar := trimmed[len(trimmed)-1]
	// 中英文标点：句号、问号、感叹号、冒号、分号
	naturalEndings := []string{"。", "？", "！", ".", "?", "!", "：", ":", "；", ";"}
	for _, ending := range naturalEndings {
		if strings.HasSuffix(trimmed, ending) {
			return true
		}
	}
	// 换行符也视为自然段落结尾
	if lastChar == '\n' {
		return true
	}
	return false
}

// ProcessCustomChunk 主函数，流式切片策略
// 策略说明：
// 1. 初始化缓冲区
// 2. 流式遍历每个 Item
// 3. 类型分支处理：
//   - 标题：不生成独立切片，只更新标题栈用于后续内容的上下文（标题内容会拼接到后续chunk中）
//   - 文本：累加到缓冲区，当 Token 超过阈值且遇到自然段落结尾时闭合生成切片
//   - 原子对象（Table/Formula/Code）：强制闭合文本切片，作为独立不可分割的原子
//   - 图片：强制闭合文本切片，提取 bbox 和路径作为独立对象
//
// elements: 可选的 Element 列表，如果提供则用于建立 Chunk 与 Element 的关联关系
func ProcessCustomChunk(ctx context.Context,
	fileName string,
	contentList []*drivenadapters.ContentItem,
	docMD5 string,
	embedding bool,
	model string,
	token string,
	elements ...[]*Element) []*CustomChunk {
	if len(contentList) == 0 {
		return []*CustomChunk{}
	}

	// 步骤0: 转换并确保基本字段
	processedList := convertToProcessingItems(contentList)
	ensureBasicFields(processedList)
	preprocessItems(processedList)

	// 建立 SegmentID 到 ElementID 的映射，并获取 documentID
	var segmentIDToElementID map[int]string
	var documentID string
	if len(elements) > 0 && len(elements[0]) > 0 {
		segmentIDToElementID = make(map[int]string)
		for i, element := range elements[0] {
			if i < len(processedList) {
				segmentIDToElementID[i] = element.ElementID
			}
		}
		// 从第一个 element 获取 documentID
		if len(elements[0]) > 0 && elements[0][0] != nil {
			documentID = elements[0][0].DocumentID
		}
	}

	var customChunks []*CustomChunk

	// ============ 流式切片状态 ============
	// 文本缓冲区
	var textBuffer []string
	var bufferPages []int
	var bufferLocations [][4]int
	var bufferSegmentIDs []int // 记录缓冲区中涉及的所有 SegmentID

	// 标题栈：维护最近的各级标题，key 是标题级别（1-3），value 是标题内容
	titleStack := make(map[int]string)

	// Token 阈值（字符数近似，中文字符约2-3 token）
	// 调整为 768 以减少切片数量，提高每个切片的信息密度
	const tokenThreshold = 768
	// 最小切片长度限制，避免生成过短的切片
	const minChunkLength = 50

	// getRelatedTitles 获取关联的标题内容（按级别从高到低）
	getRelatedTitles := func() string {
		var titles []string
		// 按级别从高到低（1, 2, 3）收集标题
		for level := 1; level <= 3; level++ {
			if title, exists := titleStack[level]; exists && title != "" {
				titles = append(titles, title)
			}
		}
		if len(titles) > 0 {
			return strings.Join(titles, "\n")
		}
		return ""
	}

	// flushBuffer 闭合当前文本缓冲区，生成一个文本切片
	// forceFlush: 是否强制输出（即使小于最小长度），用于处理文档末尾的剩余内容
	flushBuffer := func(forceFlush bool) {
		if len(textBuffer) == 0 {
			return
		}
		content := strings.Join(textBuffer, "\n")

		// 将关联的标题拼接到内容前面
		relatedTitles := getRelatedTitles()
		if relatedTitles != "" {
			content = relatedTitles + "\n" + content
		}

		// 检查最小长度限制（仅在非强制输出时检查）
		if !forceFlush && len(content) < minChunkLength {
			return
		}

		sliceMD5 := GenerateMD5(content)

		// 收集关联的 ElementIDs（缓冲区中所有 SegmentID 对应的 ElementID）
		var elementID string
		var elementIDs []string
		if segmentIDToElementID != nil && len(bufferSegmentIDs) > 0 {
			seenElementIDs := make(map[string]bool)
			for _, segID := range bufferSegmentIDs {
				if eid, ok := segmentIDToElementID[segID]; ok {
					if !seenElementIDs[eid] {
						elementIDs = append(elementIDs, eid)
						seenElementIDs[eid] = true
						// 第一个 ElementID 单独存储
						if elementID == "" {
							elementID = eid
						}
					}
				}
			}
		}

		// 使用第一个 SegmentID 作为主要标识
		firstSegmentID := -1
		if len(bufferSegmentIDs) > 0 {
			firstSegmentID = bufferSegmentIDs[0]
		}

		customChunks = append(customChunks, &CustomChunk{
			DocName:         fileName,
			DocMD5:          docMD5,
			SliceMD5:        sliceMD5,
			ID:              uuid.New().String(),
			DeduplicationID: GenerateDeduplicationID(docMD5, sliceMD5, fmt.Sprintf("%d", firstSegmentID)),
			DocumentID:      documentID,
			SliceType:       SliceTypeText,
			Pages:           uniqueSortedInts(bufferPages),
			SegmentID:       firstSegmentID,
			ElementID:       elementID,
			ElementIDs:      elementIDs,
			Location:        bufferLocations,
			SliceContent:    content,
		})

		// 重置缓冲区
		textBuffer, bufferPages, bufferLocations, bufferSegmentIDs = nil, nil, nil, nil
	}

	// createChunk 创建切片（原子对象）
	createChunk := func(item *ProcessingItem, index int, sliceType int) {
		content := item.Text

		// 代码类型优先使用 CodeBody
		if sliceType == SliceTypeCode && item.CodeBody != "" {
			content = item.CodeBody
		}
		// 表格类型使用 TableBody
		if sliceType == SliceTypeTable && item.TableBody != "" {
			content = item.TableBody
		}

		// 将关联的标题拼接到内容前面
		relatedTitles := getRelatedTitles()
		if relatedTitles != "" {
			content = relatedTitles + "\n" + content
		}

		var imgPath *string
		if sliceType == SliceTypeImage && item.ImgPath != "" {
			if url, err := UploadImageToOSS(ctx, item.ImgPath, fileName); err == nil {
				imgPath = &url
			} else {
				// 上传失败时使用原始路径
				imgPath = &item.ImgPath
			}
		}

		// 获取关联的 ElementID
		var elementID string
		var elementIDs []string
		if segmentIDToElementID != nil {
			if eid, ok := segmentIDToElementID[index]; ok {
				elementID = eid
				elementIDs = append(elementIDs, eid)
			}
		}

		sliceMD5 := GenerateMD5(content)
		customChunks = append(customChunks, &CustomChunk{
			DocName:         fileName,
			DocMD5:          docMD5,
			SliceMD5:        sliceMD5,
			ID:              uuid.New().String(),
			DeduplicationID: GenerateDeduplicationID(docMD5, sliceMD5, fmt.Sprintf("%d", index)),
			DocumentID:      documentID,
			SliceType:       sliceType,
			Pages:           extractPages(item.PageIdx),
			SegmentID:       index,
			ElementID:       elementID,
			ElementIDs:      elementIDs,
			Location:        extractLocations(item.Bbox),
			SliceContent:    content,
			ImgPath:         imgPath,
		})
	}

	// ============ 流式遍历处理 ============
	for i, item := range processedList {
		sliceType := determineSliceType(item)

		// A. 标题处理：不生成独立切片，只更新标题栈用于后续内容的上下文
		// 标题内容会通过标题栈拼接到后续的文本/原子对象chunk中
		if sliceType == SliceTypeTitle {
			flushBuffer(false) // 先闭合之前的文本缓冲区

			// 更新标题栈
			if item.TextLevel != nil && *item.TextLevel >= 1 && *item.TextLevel <= 3 {
				level := *item.TextLevel
				titleText := strings.TrimSpace(item.Text)

				// 更新当前级别的标题
				titleStack[level] = titleText

				// 清除更低级别的标题（更高级别的标题会重置文档结构）
				for l := level + 1; l <= 3; l++ {
					delete(titleStack, l)
				}
			}

			// 不生成独立的标题chunk，标题只作为上下文使用
			continue
		}

		// B. 原子对象处理（表格/公式/代码/图片）：强制闭合文本切片，作为独立不可分割的原子
		if sliceType == SliceTypeTable || sliceType == SliceTypeFormula ||
			sliceType == SliceTypeCode || sliceType == SliceTypeImage {
			flushBuffer(false)
			createChunk(item, i, sliceType)
			continue
		}

		// C. 普通文本处理：累加到缓冲区
		bufferSegmentIDs = append(bufferSegmentIDs, i)
		textBuffer = append(textBuffer, item.Text)
		bufferPages = append(bufferPages, extractPages(item.PageIdx)...)
		bufferLocations = append(bufferLocations, extractLocations(item.Bbox)...)

		// 计算当前缓冲区 Token 长度（粗略估算：字符数）
		currentLen := 0
		for _, s := range textBuffer {
			currentLen += len(s)
		}

		// 当超过阈值且遇到自然段落结尾时，闭合切片
		if currentLen > tokenThreshold && isNaturalParagraphEnd(item.Text) {
			flushBuffer(false)
		}
	}

	// 处理剩余缓冲区内容（强制输出，即使小于最小长度）
	flushBuffer(true)

	// 处理向量化
	if embedding && model != "" {
		ad := drivenadapters.NewAnyData()
		var inputs []string
		for _, chunk := range customChunks {
			inputs = append(inputs, chunk.SliceContent)
		}
		if res, err := ad.Embedding(ctx, model, inputs, token); err == nil {
			for i := range res.Data {
				if i < len(customChunks) {
					customChunks[i].Embedding = res.Data[i].Embedding
				}
			}
		}
	}

	return customChunks
}

// 辅助工具函数
func uniqueSortedInts(input []int) []int {
	if len(input) == 0 {
		return []int{}
	}
	m := make(map[int]bool)
	for _, v := range input {
		if v != -1 {
			m[v] = true
		}
	}
	var res []int
	for k := range m {
		res = append(res, k)
	}
	sort.Ints(res)
	return res
}

func extractPages(p any) []int {
	switch v := p.(type) {
	case int:
		if v == -1 {
			return []int{}
		}
		return []int{v}
	case []int:
		return v
	}
	return []int{}
}

func extractLocations(b any) [][4]int {
	switch v := b.(type) {
	case [4]int:
		return [][4]int{v}
	case [][4]int:
		return v
	}
	return [][4]int{}
}
