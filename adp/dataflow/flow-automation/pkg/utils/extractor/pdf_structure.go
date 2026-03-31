// Package extractor provides utilities for parsing PDF documents and converting them to structured element formats.
package extractor

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

// Element 目标元素结构
type Element struct {
	ElementID       string                 `json:"element_id"`
	DocumentID      string                 `json:"document_id"`
	ElementType     string                 `json:"element_type"`
	PageNo          *int                   `json:"page_no"`
	LineNo          *int                   `json:"line_no"`
	LineStart       *int                   `json:"line_start"`
	LineEnd         *int                   `json:"line_end"`
	Bbox            *BboxInfo              `json:"bbox"`
	ReadingOrder    int                    `json:"reading_order"`
	Level           int                    `json:"level"`
	ParentElementID *string                `json:"parent_element_id"`
	Content         string                 `json:"content"`
	Structure       map[string]interface{} `json:"structure"`
	Style           map[string]interface{} `json:"style"`
	ModalType       string                 `json:"modal_type"`
	Metadata        map[string]interface{} `json:"metadata"`
	ImgPath         *string                `json:"img_path,omitempty"`
}

// BboxInfo 边界框信息
type BboxInfo struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ConvertContentItemsToElements 将 ContentItem 列表转换为 Element 数组
// originalFileURL: 当输入文件本身就是图片文件时，传入原始文件的URL，用于设置img_path
func ConvertContentItemsToElements(ctx context.Context, contentList []*drivenadapters.ContentItem, documentID string, docName string, docMD5 string, originalFileURL ...string) []*Element {
	var elements []*Element
	readingOrder := 1

	// 建立索引到 element_id 的映射，用于查找父元素
	indexToElementID := make(map[int]string)

	for i, item := range contentList {
		// 生成 element_id（使用 UUID）
		elementID := uuid.New().String()
		indexToElementID[i] = elementID

		// 生成稳定的去重ID（基于docMD5和元素内容）
		elementContent := item.Text
		if item.Type == "table" && item.TableBody != "" {
			elementContent = item.TableBody
		}
		if item.Type == "code" && item.CodeBody != "" {
			elementContent = item.CodeBody
		}
		// image 类型需要包含 ImageCaption 和 ImageFootnote 用于去重
		if item.Type == "image" {
			imageParts := []string{}
			if strings.TrimSpace(item.Text) != "" {
				imageParts = append(imageParts, strings.TrimSpace(item.Text))
			}
			if len(item.ImageCaption) > 0 {
				caption := strings.Join(item.ImageCaption, "\n")
				if strings.TrimSpace(caption) != "" {
					imageParts = append(imageParts, strings.TrimSpace(caption))
				}
			}
			if len(item.ImageFootnote) > 0 {
				footnote := strings.Join(item.ImageFootnote, "\n")
				if strings.TrimSpace(footnote) != "" {
					imageParts = append(imageParts, strings.TrimSpace(footnote))
				}
			}
			if len(imageParts) > 0 {
				elementContent = strings.Join(imageParts, "\n")
			}
		}
		contentHash := GenerateMD5(elementContent)
		deduplicationID := GenerateDeduplicationID(docMD5, fmt.Sprintf("%d", i), contentHash[:8], item.Type, fmt.Sprintf("%d", item.PageIdx))

		// 映射 element_type
		elementType := MapContentItemTypeToElementType(item.Type)

		// 转换 bbox
		var bbox *BboxInfo
		if len(item.Bbox) == 4 {
			x := float64(item.Bbox[0])
			y := float64(item.Bbox[1])
			width := float64(item.Bbox[2] - item.Bbox[0])
			height := float64(item.Bbox[3] - item.Bbox[1])
			bbox = &BboxInfo{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
			}
		}

		// 获取页码
		var pageNo *int
		if item.PageIdx >= 0 {
			pageNoVal := item.PageIdx + 1 // PDF页码从1开始
			pageNo = &pageNoVal
		}

		// 获取 level
		level := 0
		if item.TextLevel != nil {
			level = *item.TextLevel
		}

		// 查找父元素（查找最近的 level 更小的元素）
		var parentElementID *string
		for j := i - 1; j >= 0; j-- {
			prevItem := contentList[j]
			var prevLevel int
			if prevItem.TextLevel != nil {
				prevLevel = *prevItem.TextLevel
			}
			if prevLevel < level {
				// 找到父元素，从映射中获取其 element_id
				if parentID, ok := indexToElementID[j]; ok {
					parentElementID = &parentID
				}
				break
			}
		}

		// 构建 structure
		structure := make(map[string]interface{})
		if item.Type == "table" && item.TableBody != "" {
			structure["html"] = item.TableBody
		}

		// 构建 style
		style := make(map[string]interface{})

		// 设置 img_path（仅当元素类型为 image 时）
		var imgPath *string
		if item.Type == "image" {
			// 如果提供了原始文件URL（说明输入文件本身就是图片），需要上传到OSS
			if len(originalFileURL) > 0 && originalFileURL[0] != "" {
				// 从原始文件URL下载并上传到OSS，获取OSS的下载链接
				uploadedURL, err := UploadOriginalImageToOSS(ctx, originalFileURL[0], docName)
				if err != nil {
					// 如果上传失败，记录警告但不设置img_path
					traceLog.WithContext(ctx).Warnf("[convertContentItemsToElements] failed to upload original image: %s, error: %v", originalFileURL[0], err)
				} else {
					imgPath = &uploadedURL
				}
			} else if item.ImgPath != "" {
				// 否则使用解析出的图片路径，上传到 OSS 并获取下载链接
				uploadedURL, err := UploadImageToOSS(ctx, item.ImgPath, docName)
				if err != nil {
					// 如果上传失败，记录警告但继续使用原始路径
					traceLog.WithContext(ctx).Warnf("[convertContentItemsToElements] failed to upload image: %s, error: %v", item.ImgPath, err)
					imgPath = &item.ImgPath
				} else {
					imgPath = &uploadedURL
				}
			}
		}

		// 设置 Content，根据类型选择合适的内容
		content := item.Text
		if item.Type == "code" && item.CodeBody != "" {
			content = item.CodeBody
		}
		// list 类型使用 ListItems 合并为文本
		if item.Type == "list" && len(item.ListItems) > 0 {
			listContent := strings.Join(item.ListItems, "\n")
			if listContent != "" {
				content = listContent
			}
		}
		// image 类型聚合 ImageCaption 和 ImageFootnote 到 content
		if item.Type == "image" {
			textParts := []string{}
			if strings.TrimSpace(item.Text) != "" {
				textParts = append(textParts, strings.TrimSpace(item.Text))
			}
			if len(item.ImageCaption) > 0 {
				caption := strings.Join(item.ImageCaption, "\n")
				if strings.TrimSpace(caption) != "" {
					textParts = append(textParts, strings.TrimSpace(caption))
				}
			}
			if len(item.ImageFootnote) > 0 {
				footnote := strings.Join(item.ImageFootnote, "\n")
				if strings.TrimSpace(footnote) != "" {
					textParts = append(textParts, strings.TrimSpace(footnote))
				}
			}
			// 去重并合并
			uniqueParts := []string{}
			seen := make(map[string]bool)
			for _, part := range textParts {
				if part != "" && !seen[part] {
					uniqueParts = append(uniqueParts, part)
					seen[part] = true
				}
			}
			if len(uniqueParts) > 0 {
				content = strings.Join(uniqueParts, "\n")
			}
		}

		// 创建 Element
		element := &Element{
			ElementID:       elementID,
			DocumentID:      documentID,
			ElementType:     elementType,
			PageNo:          pageNo,
			LineNo:          nil, // PDF文档通常没有行号
			LineStart:       nil,
			LineEnd:         nil,
			Bbox:            bbox,
			ReadingOrder:    readingOrder,
			Level:           level,
			ParentElementID: parentElementID,
			Content:         content,
			Structure:       structure,
			Style:           style,
			ModalType:       "text", // 默认为text
			Metadata: map[string]interface{}{
				"source_format":    "pdf",
				"deduplication_id": deduplicationID, // 添加去重标识
			},
			ImgPath: imgPath,
		}

		readingOrder++
		elements = append(elements, element)
	}

	return elements
}

// MapContentItemTypeToElementType 将 ContentItem 的 type 映射为 Element 的 element_type
func MapContentItemTypeToElementType(contentItemType string) string {
	switch contentItemType {
	case "title":
		return "heading"
	case "text":
		return "text"
	case "table":
		return "table"
	case "image":
		return "figure"
	case "list":
		return "list"
	case "code":
		return "code_block"
	case "header":
		return "header_footer"
	case "footer":
		return "header_footer"
	case "footnote":
		return "footnote"
	case "separator":
		return "separator"
	case "formula", "equation":
		return "formula"
	case "link":
		return "link"
	default:
		return "text"
	}
}
