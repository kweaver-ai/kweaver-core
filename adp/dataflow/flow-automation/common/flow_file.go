package common

import (
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
)

const (
	// DFSURIPrefix dfs:// 协议前缀
	DFSURIPrefix = "dfs://"
	// FlowFileObjectPrefix Dataflow 文件对象存储路径前缀
	FlowFileObjectPrefix = "dataflow_files"
)

var (
	// ErrInvalidDFSURI 无效的 dfs:// URI
	ErrInvalidDFSURI = errors.New("invalid dfs:// uri")
	// ErrInvalidDFSURIPrefix 无效的 dfs:// 协议前缀
	ErrInvalidDFSURIPrefix = errors.New("invalid dfs:// uri prefix")
	// ErrInvalidFileID 无效的文件 ID
	ErrInvalidFileID = errors.New("invalid file id")
	// ErrInvalidFileName 无效的文件名
	ErrInvalidFileName = errors.New("invalid file name")
)

// BuildDFSURI 生成 dfs://<file_id> 格式的 URI
func BuildDFSURI(fileID uint64) string {
	return fmt.Sprintf("%s%d", DFSURIPrefix, fileID)
}

// ParseDFSURI 解析 dfs://<file_id> 格式的 URI，返回 file_id
func ParseDFSURI(uri string) (uint64, error) {
	if uri == "" {
		return 0, ErrInvalidDFSURI
	}

	if !strings.HasPrefix(uri, DFSURIPrefix) {
		return 0, ErrInvalidDFSURIPrefix
	}

	fileIDStr := strings.TrimPrefix(uri, DFSURIPrefix)
	if fileIDStr == "" {
		return 0, ErrInvalidFileID
	}

	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		return 0, ErrInvalidFileID
	}

	return fileID, nil
}

// IsDFSURI 判断是否为有效的 dfs:// URI
func IsDFSURI(uri string) bool {
	if uri == "" {
		return false
	}
	return strings.HasPrefix(uri, DFSURIPrefix)
}

// NormalizeFileID 归一化文件 ID，支持 dfs://<id> 或纯 ID
// 如果输入是 dfs://<id>，返回解析后的 id
// 如果输入是纯数字，直接返回
func NormalizeFileID(input string) (uint64, error) {
	if input == "" {
		return 0, ErrInvalidFileID
	}

	if IsDFSURI(input) {
		return ParseDFSURI(input)
	}

	fileID, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		return 0, ErrInvalidFileID
	}

	return fileID, nil
}

// ============================================================
// 对象存储路径生成
// ============================================================

// BuildFlowFileObjectKey 生成 Dataflow 文件对象的 OSS 存储路径
// fileID: flow_file 表的主键 ID
// filename: 原始文件名
func BuildFlowFileObjectKey(fileID uint64, filename string) (string, error) {

	if filename == "" {
		return "", ErrInvalidFileName
	}

	// 清理文件名，防止路径穿越
	cleanName := path.Base(filename)
	if cleanName == "." || cleanName == ".." {
		return "", ErrInvalidFileName
	}

	config := NewConfig()
	prefix := strings.Trim(config.Server.StoragePrefix, "/")

	if prefix != "" {
		return fmt.Sprintf("%s/%s/%d/%s", prefix, FlowFileObjectPrefix, fileID, cleanName), nil
	}

	return fmt.Sprintf("%s/%d/%s", FlowFileObjectPrefix, fileID, cleanName), nil
}
