package mgnt

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/s3"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/log"
)

// S3DataItem S3数据项
type S3DataItem struct {
	ID           string `json:"id"`            // 唯一标识符 (bucket/key的组合)
	Bucket       string `json:"bucket"`        // 存储桶名称
	Key          string `json:"key"`           // 对象键(路径)
	Name         string `json:"name"`          // 文件名
	Size         int64  `json:"size"`          // 文件大小(字节)
	LastModified string `json:"last_modified"` // 最后修改时间
	ETag         string `json:"etag"`          // ETag (通常是MD5值)
	DownloadURL  string `json:"download_url"`  // 下载链接
}

// ValidateS3Config 验证S3配置
func (m *mgnt) ValidateS3Config(ctx context.Context, bucket, path, mode string) (*S3ValidationResult, error) {
	result := &S3ValidationResult{}

	// 检查S3适配器是否已初始化
	if m.s3Adapter == nil {
		return nil, fmt.Errorf("S3 adapter not initialized, please configure S3 credentials")
	}

	// 验证模式
	if mode != "upload" {
		return nil, fmt.Errorf("unsupported mode: %s", mode)
	}

	// 使用libs/go/s3的Bucket
	conn := s3.NewS3().GetDefaultConnection()
	if conn == nil {
		return nil, fmt.Errorf("system S3 connection not configured")
	}
	sysBucket := conn.BucketName

	// 验证Bucket
	if err := m.s3Adapter.ValidateBucket(ctx, sysBucket); err != nil {
		result.BucketExists = false
		result.Message = fmt.Sprintf("System bucket validation failed: %v", err)
		log.Warnf("[mgnt.ValidateS3Config] System bucket validation failed: %v", err)
		return result, nil
	}
	result.BucketExists = true

	// upload模式下，不再验证路径下的文件数量，直接返回成功
	// 因为文件是由用户上传的，路径固定
	result.PathAccessible = true
	result.Message = "Successfully validated system bucket"

	log.Infof("[mgnt.ValidateS3Config] Validation successful for bucket=%s (system default), mode=%s", sysBucket, mode)
	return result, nil
}

// UploadS3File 上传S3文件
func (m *mgnt) UploadS3File(ctx context.Context, dagID string, fileHeader *multipart.FileHeader, userInfo *drivenadapters.UserInfo) (*S3DataItem, error) {
	// 1. 验证系统S3配置
	conn := s3.NewS3().GetDefaultConnection()
	if conn == nil {
		return nil, fmt.Errorf("system S3 connection not configured")
	}
	sysBucket := conn.BucketName

	if m.s3Adapter == nil {
		return nil, fmt.Errorf("S3 adapter not initialized")
	}

	// 2. 验证文件
	// 这里可以添加文件类型、大小限制等验证
	// 例如：if fileHeader.Size > 100*1024*1024 { return nil, fmt.Errorf("file too large") }

	// 3. 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// 4. 生成对象Key
	// 格式: dataflow-doc/{dag_id}/{filename}
	// 如果dagID是临时ID（例如长度不等于18位数字），则存储在临时目录: dataflow-doc/temp/{dag_id}/{filename}
	// 注意：文件名需要清洗，防止路径遍历等注入
	filename := filepath.Base(fileHeader.Filename)
	var key string
	if len(dagID) != 18 { // 简单假设，非正式ID都当做临时ID处理，或者可以判断由前端传来的特定前缀
		key = fmt.Sprintf("dataflow-doc/temp/%s/%s", dagID, filename)
	} else {
		key = fmt.Sprintf("dataflow-doc/%s/%s", dagID, filename)
	}

	// 5. 上传文件
	err = m.s3Adapter.Upload(ctx, sysBucket, key, file)
	if err != nil {
		log.Errorf("[mgnt.UploadS3File] Upload failed: %v", err)
		return nil, fmt.Errorf("upload failed: %v", err)
	}

	// 6. 生成预签名URL (用于返回给前端回显)
	downloadURL, err := m.s3Adapter.GeneratePresignedURL(ctx, sysBucket, key, 7*24*time.Hour)
	if err != nil {
		log.Warnf("[mgnt.UploadS3File] GeneratePresignedURL failed: %v", err)
		// fallback
		downloadURL = fmt.Sprintf("%s/%s/%s", conn.Endpoint, sysBucket, key)
	}

	// 7. 返回结果
	item := &S3DataItem{
		ID:           fmt.Sprintf("%s/%s", sysBucket, key),
		Bucket:       sysBucket,
		Key:          key,
		Name:         filename,
		Size:         fileHeader.Size,
		LastModified: time.Now().Format("2006-01-02T15:04:05Z"), // 或者是S3返回的时间
		DownloadURL:  downloadURL,
	}

	log.Infof("[mgnt.UploadS3File] File uploaded successfully: %s, size: %d, user: %s", key, fileHeader.Size, userInfo.UserID)
	return item, nil
}

// ListS3Files 列举S3文件
func (m *mgnt) ListS3Files(ctx context.Context, dagID string, userInfo *drivenadapters.UserInfo) ([]*S3DataItem, error) {
	// 1. 验证系统S3配置
	conn := s3.NewS3().GetDefaultConnection()
	if conn == nil {
		return nil, fmt.Errorf("system S3 connection not configured")
	}
	sysBucket := conn.BucketName

	if m.s3Adapter == nil {
		return nil, fmt.Errorf("S3 adapter not initialized")
	}

	// 2. 列举件
	var prefix string
	if len(dagID) != 18 {
		prefix = fmt.Sprintf("dataflow-doc/temp/%s/", dagID)
	} else {
		prefix = fmt.Sprintf("dataflow-doc/%s/", dagID)
	}

	objects, err := m.s3Adapter.ListObjects(ctx, sysBucket, prefix)
	if err != nil {
		log.Errorf("[mgnt.ListS3Files] ListObjects failed: %v", err)
		return nil, fmt.Errorf("failed to list objects: %v", err)
	}

	// 3. 转换结果
	var items []*S3DataItem
	for _, obj := range objects {
		// 过滤掉目录本身（如果有的话，S3通常没有）
		if obj.Key == prefix {
			continue
		}

		downloadURL, err := m.s3Adapter.GeneratePresignedURL(ctx, sysBucket, obj.Key, 7*24*time.Hour)
		if err != nil {
			log.Warnf("[mgnt.ListS3Files] GeneratePresignedURL failed for %s: %v", obj.Key, err)
			downloadURL = fmt.Sprintf("%s/%s/%s", conn.Endpoint, sysBucket, obj.Key)
		}

		items = append(items, &S3DataItem{
			ID:           fmt.Sprintf("%s/%s", sysBucket, obj.Key),
			Bucket:       sysBucket,
			Key:          obj.Key,
			Name:         filepath.Base(obj.Key),
			Size:         obj.Size,
			LastModified: obj.LastModified.Format("2006-01-02T15:04:05Z"),
			ETag:         strings.Trim(obj.ETag, "\""),
			DownloadURL:  downloadURL,
		})
	}

	return items, nil
}

// DeleteS3File 删除S3文件
func (m *mgnt) DeleteS3File(ctx context.Context, dagID, key string, userInfo *drivenadapters.UserInfo) error {
	// 1. 验证系统S3配置
	conn := s3.NewS3().GetDefaultConnection()
	if conn == nil {
		return fmt.Errorf("system S3 connection not configured")
	}
	sysBucket := conn.BucketName

	if m.s3Adapter == nil {
		return fmt.Errorf("S3 adapter not initialized")
	}

	// 2. 验证权限/Key合法性
	// 必须属于该DAG目录 dataflow-doc/{dagID}/ 或 dataflow-doc/temp/{dagID}/
	prefix := fmt.Sprintf("dataflow-doc/%s/", dagID)
	tempPrefix := fmt.Sprintf("dataflow-doc/temp/%s/", dagID)

	if !strings.HasPrefix(key, prefix) && !strings.HasPrefix(key, tempPrefix) {
		return fmt.Errorf("invalid file key: %s, must start with %s or %s", key, prefix, tempPrefix)
	}

	// 3. 删除文件
	err := m.s3Adapter.DeleteObject(ctx, sysBucket, key)
	if err != nil {
		log.Errorf("[mgnt.DeleteS3File] DeleteObject failed: %s, err: %v", key, err)
		return fmt.Errorf("failed to delete object: %v", err)
	}

	log.Infof("[mgnt.DeleteS3File] File deleted successfully: %s, user: %s", key, userInfo.UserID)
	return nil
}

// MoveS3Files 移动S3文件或目录 (用于从临时目录移动到正式目录)
// sources: key或path的列表
func (m *mgnt) MoveS3Files(ctx context.Context, sources []string, targetDagID string) ([]string, error) {
	conn := s3.NewS3().GetDefaultConnection()
	if conn == nil {
		return nil, fmt.Errorf("system S3 connection not configured")
	}
	sysBucket := conn.BucketName

	if m.s3Adapter == nil {
		return nil, fmt.Errorf("S3 adapter not initialized")
	}

	var newKeys []string

	for _, sourcePath := range sources {
		// 检查是否在temp目录下
		if !strings.Contains(sourcePath, "/temp/") {
			newKeys = append(newKeys, sourcePath)
			continue
		}

		// 1. 列举该路径下的所有对象 (支持目录移动)
		objects, err := m.s3Adapter.ListObjects(ctx, sysBucket, sourcePath)
		if err != nil {
			log.Errorf("[mgnt.MoveS3Files] ListObjects failed for path %s: %v", sourcePath, err)
			return nil, fmt.Errorf("failed to list objects for path %s: %v", sourcePath, err)
		}

		// 2. 遍历移动每个对象
		for _, obj := range objects {
			sourceKey := obj.Key

			// 解析源Key结构，提取SessionID并构造目标Key
			// 假设结构为 dataflow-doc/temp/{sessionID}/...
			// 目标结构 dataflow-doc/{dagID}/...
			parts := strings.Split(sourceKey, "/")
			if len(parts) < 3 || parts[1] != "temp" {
				log.Warnf("[mgnt.MoveS3Files] Skipping invalid temp key structure: %s", sourceKey)
				continue
			}
			sessionID := parts[2]
			tempPrefix := fmt.Sprintf("dataflow-doc/temp/%s", sessionID)
			targetPrefix := fmt.Sprintf("dataflow-doc/%s", targetDagID)

			targetKey := strings.Replace(sourceKey, tempPrefix, targetPrefix, 1)

			// 2.1 Copy
			err := m.s3Adapter.CopyObject(ctx, sysBucket, sourceKey, targetKey)
			if err != nil {
				log.Errorf("[mgnt.MoveS3Files] CopyObject failed: %v", err)
				return nil, fmt.Errorf("failed to copy object %s: %v", sourceKey, err)
			}

			// 2.2 Delete source
			err = m.s3Adapter.DeleteObject(ctx, sysBucket, sourceKey)
			if err != nil {
				log.Warnf("[mgnt.MoveS3Files] DeleteObject (source) failed: %v", err)
				// don't fail, just warn
			}

			log.Infof("[mgnt.MoveS3Files] Moved %s to %s", sourceKey, targetKey)
		}

		// 3. 计算并返回新的Path (保持输入输出一一对应)
		// 我们对输入路径 sourcePath 进行同样的路径替换
		parts := strings.Split(sourcePath, "/")
		if len(parts) >= 3 && parts[1] == "temp" {
			sessionID := parts[2]
			tempPrefix := fmt.Sprintf("dataflow-doc/temp/%s", sessionID)
			targetPrefix := fmt.Sprintf("dataflow-doc/%s", targetDagID)
			newPath := strings.Replace(sourcePath, tempPrefix, targetPrefix, 1)
			newKeys = append(newKeys, newPath)
		} else {
			// Fallback: 如果无法解析，保留原值 (理论上不应发生，因为前面filter了)
			newKeys = append(newKeys, sourcePath)
		}
	}

	return newKeys, nil
}

// handleS3DataSource 处理S3数据源
func (m *mgnt) handleS3DataSource(ctx context.Context, dag *entity.Dag, dagIns *entity.DagInstance) error {
	if dag.TriggerConfig == nil || dag.TriggerConfig.DataSource == nil {
		return fmt.Errorf("trigger config or data source is nil")
	}

	dataSource := dag.TriggerConfig.DataSource
	if dataSource.Operator != common.S3DataListObjects {
		return fmt.Errorf("unsupported data source operator: %s", dataSource.Operator)
	}

	// 解析S3配置参数
	if dataSource.Parameters == nil {
		return fmt.Errorf("data source parameters is nil")
	}

	// 获取系统配置的Bucket
	conn := s3.NewS3().GetDefaultConnection()
	if conn == nil {
		return fmt.Errorf("system S3 connection not configured")
	}
	sysBucket := conn.BucketName

	// 获取S3适配器
	s3Adapter := m.s3Adapter
	if s3Adapter == nil {
		return fmt.Errorf("S3 adapter not initialized")
	}

	mode := dataSource.Parameters.Mode
	if mode == "" {
		mode = "upload" // 默认为upload
	}

	if mode != "upload" {
		return fmt.Errorf("unsupported mode: %s", mode)
	}

	// Upload模式下，路径固定为 dataflow-doc/{dag_id}
	// 但实际上，在Upload模式下，前端传递的是sources列表，其中包含了文件的key
	// 我们直接使用sources中的key即可，不需要再次列举，或者为了安全起见，我们根据dagID列举并过滤

	// 这里我们遵循设计：TriggerConfig中包含sources列表 (由前端在点击执行/保存时生成? 或者我们直接扫描目录?)
	// 逻辑设计中: 2.3.1 支持上传文件... 路径固定...
	// 3.2.1 S3数据源参数... sources array... sources[].key

	// 如果Parameters.Sources不为空，则直接使用其中的Key (前端传递)
	// 但为了防止越权，我们需要验证这些Key是否属于 dataflow-doc/{dag_id}/ 前缀

	prefix := fmt.Sprintf("dataflow-doc/%s/", dag.ID)

	sources := dataSource.Parameters.Sources
	var allItems []S3DataItem

	if len(sources) > 0 {

		// 使用前端传递的文件列表
		for _, source := range sources {
			var effectiveKey string
			if source.Key != "" {
				effectiveKey = source.Key
			} else if source.Path != "" {
				effectiveKey = source.Path
			}

			if effectiveKey == "" {
				continue
			}

			// 验证前缀
			if !strings.HasPrefix(effectiveKey, prefix) && !strings.HasPrefix(effectiveKey, "dataflow-doc/temp/"+dag.ID) { // Also allow temp path if needed, but strict prefix check requested?
				// The original code passed `prefix` which was `dataflow-doc/{dag.ID}/`.
				log.Warnf("[mgnt.handleS3DataSource] Key %s does not match prefix %s, skipping", effectiveKey, prefix)
				continue
			}

			// 验证Bucket (必须是系统Bucket)
			// source.Bucket 可能是前端传的，也可能没有。我们强制使用系统Bucket

			// 生成下载链接
			id := fmt.Sprintf("%s/%s", sysBucket, effectiveKey)
			downloadURL, err := s3Adapter.GeneratePresignedURL(ctx, sysBucket, effectiveKey, 7*24*time.Hour)
			if err != nil {
				log.Warnf("[mgnt.handleS3DataSource] Failed to generate presigned URL for %s: %v", effectiveKey, err)
				downloadURL = fmt.Sprintf("%s/%s/%s", conn.Endpoint, sysBucket, effectiveKey)
			}

			// 获取对象元数据
			meta, err := s3Adapter.GetObjectMetadata(ctx, sysBucket, effectiveKey)
			var size int64
			var lastModified string
			var etag string

			if err != nil {
				log.Warnf("[mgnt.handleS3DataSource] Failed to get object metadata for %s: %v", effectiveKey, err)
				// Fallback to provided size if available
				size = source.Size
			} else {
				size = meta.ContentLength
				lastModified = meta.LastModified.Format(time.RFC3339)
				etag = strings.Trim(meta.ETag, "\"")
			}

			allItems = append(allItems, S3DataItem{
				ID:           id,
				Bucket:       sysBucket,
				Key:          effectiveKey,
				Name:         source.Name, // Use typed Name field
				Size:         size,
				LastModified: lastModified,
				ETag:         etag,
				DownloadURL:  downloadURL,
			})
		}
	} else {
		// 如果没有sources，我们要扫描目录下所有文件吗？
		// 逻辑设计中说 "支持列举文件... 本期仅支持upload模式"。
		// 通常upload模式意味着处理用户上传的那些文件。
		// 如果DataFlow被触发，应该是针对某些文件的。
		// 如果是手动触发，可能带有参数。
		// 这里假设如果没有sources，就扫描整个目录

		objects, err := s3Adapter.ListObjects(ctx, sysBucket, prefix)
		if err != nil {
			return fmt.Errorf("failed to list objects: %v", err)
		}

		for _, obj := range objects {
			id := fmt.Sprintf("%s/%s", sysBucket, obj.Key)
			downloadURL, err := s3Adapter.GeneratePresignedURL(ctx, sysBucket, obj.Key, 7*24*time.Hour)
			if err != nil {
				downloadURL = fmt.Sprintf("%s/%s/%s", conn.Endpoint, sysBucket, obj.Key)
			}

			allItems = append(allItems, S3DataItem{
				ID:           id,
				Bucket:       sysBucket,
				Key:          obj.Key,
				Name:         filepath.Base(obj.Key),
				Size:         obj.Size,
				LastModified: obj.LastModified.Format(time.RFC3339),
				ETag:         strings.Trim(obj.ETag, "\""),
				DownloadURL:  downloadURL,
			})
		}
	}

	if len(allItems) == 0 {
		log.Warnf("[mgnt.handleS3DataSource] no objects found for dag %s", dag.ID)
		// return fmt.Errorf("no objects found") // 也许不应该报错，而是空列表
	}

	// 将S3对象列表存储到共享数据中
	// 触发器步骤可以访问 __0 来获取数据项列表
	dagIns.ShareData.Set("0", map[string]interface{}{
		"items": allItems,
		"count": len(allItems),
	})

	log.Infof("[mgnt.handleS3DataSource] Successfully loaded %d objects from S3 for dag %s", len(allItems), dag.ID)
	return nil
}

// GetS3FileDownloadURL 获取S3文件下载链接
func (m *mgnt) GetS3FileDownloadURL(ctx context.Context, dagID, key string, userInfo *drivenadapters.UserInfo) (string, error) {
	// 1. 验证系统S3配置
	conn := s3.NewS3().GetDefaultConnection()
	if conn == nil {
		return "", fmt.Errorf("system S3 connection not configured")
	}
	sysBucket := conn.BucketName

	if m.s3Adapter == nil {
		return "", fmt.Errorf("S3 adapter not initialized")
	}

	// 2. 验证权限/Key合法性
	// 必须属于该DAG目录 dataflow-doc/{dagID}/ 或 dataflow-doc/temp/{dagID}/
	prefix := fmt.Sprintf("dataflow-doc/%s/", dagID)
	tempPrefix := fmt.Sprintf("dataflow-doc/temp/%s/", dagID)

	if !strings.HasPrefix(key, prefix) && !strings.HasPrefix(key, tempPrefix) {
		return "", fmt.Errorf("invalid file key: %s, must start with %s or %s", key, prefix, tempPrefix)
	}

	// 3. 生成预签名URL (有效期1小时)
	url, err := m.s3Adapter.GeneratePresignedURL(ctx, sysBucket, key, 1*time.Hour)
	if err != nil {
		log.Errorf("[mgnt.GetS3FileDownloadURL] GeneratePresignedURL failed: %s, err: %v", key, err)
		return "", fmt.Errorf("failed to generate download url: %v", err)
	}

	log.Infof("[mgnt.GetS3FileDownloadURL] Download URL generated for: %s, user: %s", key, userInfo.UserID)
	return url, nil
}
