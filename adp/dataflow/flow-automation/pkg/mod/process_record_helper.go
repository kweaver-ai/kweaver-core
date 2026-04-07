package mod

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/store"
)

// ExtractObjectKey 从DagInstance中提取处理对象标识
// 对于文档，使用 docID + rev（优先）或 docID + md5 作为唯一标识
// 返回格式：docID:rev 或 docID:md5 或 docID（如果rev和md5都不存在）
func ExtractObjectKey(dagIns *entity.DagInstance) (string, error) {
	if dagIns == nil {
		return "", fmt.Errorf("dagIns is nil")
	}

	var docID string
	var rev string
	var md5Value string

	// 优先从 vars.source 中提取
	if sourceVar, ok := dagIns.Vars["source"]; ok {
		sourceStr := sourceVar.Value
		if sourceStr != "" {
			// 解析 source JSON
			var sourceData map[string]interface{}
			if err := json.Unmarshal([]byte(sourceStr), &sourceData); err == nil {
				// 提取 id 字段
				if id, ok := sourceData["id"].(string); ok && id != "" {
					docID = id
				}
				// 提取 rev 字段（优先使用）
				if revVal, ok := sourceData["rev"].(string); ok && revVal != "" {
					rev = revVal
				}
				// 提取 md5 字段（如果rev不存在）
				if md5Val, ok := sourceData["md5"].(string); ok && md5Val != "" {
					md5Value = md5Val
				}
			}
		}
	}

	// 如果从source中没获取到docID，尝试从其他字段获取
	if docID == "" {
		if docIDVar, ok := dagIns.Vars["docid"]; ok {
			if docIDVal := docIDVar.Value; docIDVal != "" {
				docID = docIDVal
			}
		}
		if docID == "" {
			if docIDVar, ok := dagIns.Vars["doc_id"]; ok {
				if docIDVal := docIDVar.Value; docIDVal != "" {
					docID = docIDVal
				}
			}
		}
	}

	if docID == "" {
		return "", fmt.Errorf("no docID found in dagIns vars")
	}

	// 优先使用 rev，如果不存在则使用 md5，都不存在则只使用 docID
	if rev != "" {
		return fmt.Sprintf("%s:%s", docID, rev), nil
	}

	// 尝试从 ShareData 中获取 docMD5（ContentFileParse 可能存储在这里）
	if dagIns.ShareData != nil {
		if docMD5Val, ok := dagIns.ShareData.Get("docMD5"); ok {
			if docMD5Str, ok := docMD5Val.(string); ok && docMD5Str != "" {
				md5Value = docMD5Str
			}
		}
		// 也尝试从 elements 或 chunks 中提取 docMD5
		if elementsVal, ok := dagIns.ShareData.Get("elements"); ok {
			if elementsStr, ok := elementsVal.(string); ok {
				var elements []map[string]interface{}
				if err := json.Unmarshal([]byte(elementsStr), &elements); err == nil && len(elements) > 0 {
					if docMD5, ok := elements[0]["doc_md5"].(string); ok && docMD5 != "" {
						md5Value = docMD5
					}
				}
			}
		}
	}

	if md5Value != "" {
		return fmt.Sprintf("%s:%s", docID, md5Value), nil
	}

	// 如果 rev 和 md5 都不存在，只返回 docID（兼容旧逻辑）
	return docID, nil
}

// generateProcessCacheHash 生成处理缓存的hash值
// hash = MD5(objectKey + dagID + versionID)
// objectKey 格式：docID:rev 或 docID:md5 或 docID
func generateProcessCacheHash(objectKey, dagID, versionID string) string {
	combined := fmt.Sprintf("%s:%s:%s", objectKey, dagID, versionID)
	hash := md5.Sum([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// FindProcessedDagInstance 查询t_task_cache表中是否存在已成功处理的相同DagInstance
// 使用hash = MD5(objectKey + dagID + versionID)作为唯一标识
// objectKey 格式：docID:rev 或 docID:md5 或 docID
// 返回找到的已执行DagInstanceID（存储在f_err_msg字段中），如果未找到返回空字符串
func FindProcessedDagInstance(ctx context.Context, dagIns *entity.DagInstance, objectKey string) (string, error) {
	if dagIns == nil || objectKey == "" {
		return "", fmt.Errorf("invalid parameters: dagIns or objectKey is empty")
	}

	// 如果VersionID为空，无法进行去重检查
	if dagIns.VersionID == "" {
		return "", nil
	}

	log := traceLog.WithContext(ctx)

	// 生成hash：MD5(docID + dagID + versionID)
	hash := generateProcessCacheHash(objectKey, dagIns.DagID, dagIns.VersionID)

	// 查询t_task_cache表
	taskCache := rds.GetTaskCache()
	cacheItem, err := taskCache.GetByHash(ctx, hash)
	if err != nil {
		log.Warnf("[FindProcessedDagInstance] GetByHash failed: %s", err.Error())
		return "", fmt.Errorf("failed to query task cache: %w", err)
	}

	// 如果找到且状态为成功，返回已执行的DagInstanceID（存储在ErrMsg字段中）
	if cacheItem != nil && cacheItem.Status == rds.TaskStatusSuccess && cacheItem.Type == "dag_process" {
		processedDagInsID := cacheItem.ErrMsg // 复用ErrMsg字段存储processedDagInsID
		if processedDagInsID != "" {
			log.Infof("[FindProcessedDagInstance] Found processed dag instance in cache: hash=%s, dagInsID=%s, dagID=%s, versionID=%s, objectKey=%s",
				hash, processedDagInsID, dagIns.DagID, dagIns.VersionID, objectKey)
			return processedDagInsID, nil
		}
	}

	return "", nil
}

// SaveProcessedDagInstance 保存已处理的DagInstance到t_task_cache表
// hash = MD5(objectKey + dagID + versionID)
// objectKey 格式：docID:rev 或 docID:md5 或 docID
// f_type = "dag_process"
// f_status = TaskStatusSuccess
// f_err_msg = processedDagInsID（存储已执行的DagInstanceID）
func SaveProcessedDagInstance(ctx context.Context, objectKey, dagID, versionID, processedDagInsID string) error {
	if objectKey == "" || dagID == "" || versionID == "" || processedDagInsID == "" {
		return fmt.Errorf("invalid parameters: all fields are required")
	}

	log := traceLog.WithContext(ctx)

	// 生成hash
	hash := generateProcessCacheHash(objectKey, dagID, versionID)

	// 检查是否已存在
	taskCache := rds.GetTaskCache()
	existingItem, err := taskCache.GetByHash(ctx, hash)
	if err != nil {
		log.Warnf("[SaveProcessedDagInstance] GetByHash failed: %s", err.Error())
		// 继续执行插入，如果已存在会在Insert时失败
	}

	now := time.Now().Unix()
	if existingItem != nil {
		// 如果已存在，更新状态和processedDagInsID
		updateItem := &rds.TaskCacheItem{
			Hash:       hash,
			Status:     rds.TaskStatusSuccess,
			ErrMsg:     processedDagInsID,
			ModifyTime: now,
		}
		if err := taskCache.Update(ctx, updateItem); err != nil {
			log.Warnf("[SaveProcessedDagInstance] Update failed: %s", err.Error())
			return fmt.Errorf("failed to update task cache: %w", err)
		}
		log.Infof("[SaveProcessedDagInstance] Updated processed dag instance cache: hash=%s, dagInsID=%s", hash, processedDagInsID)
	} else {
		// 如果不存在，插入新记录
		cacheItem := &rds.TaskCacheItem{
			ID:         store.NextID(), // 使用store.NextID()生成唯一ID
			Hash:       hash,
			Type:       "dag_process",
			Status:     rds.TaskStatusSuccess,
			ErrMsg:     processedDagInsID, // 复用ErrMsg字段存储processedDagInsID
			CreateTime: now,
			ModifyTime: now,
			ExpireTime: 0, // 不过期，或根据需要设置过期时间
		}
		if err := taskCache.Insert(ctx, cacheItem); err != nil {
			log.Warnf("[SaveProcessedDagInstance] Insert failed: %s", err.Error())
			return fmt.Errorf("failed to insert task cache: %w", err)
		}
		log.Infof("[SaveProcessedDagInstance] Saved processed dag instance cache: hash=%s, dagInsID=%s", hash, processedDagInsID)
	}

	return nil
}
