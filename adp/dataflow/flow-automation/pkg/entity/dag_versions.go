package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// 定义语义版本的正则表达式模式
var semverRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z\-\.]+))?(?:\+([0-9A-Za-z\-\.]+))?$`)

// DagVersion 数据流版本
type DagVersion struct {
	BaseInfo  `yaml:",inline" json:",inline" bson:"inline"`
	DagID     string  `yaml:"dagId" json:"dagId" bson:"dagId"`
	UserID    string  `yaml:"userid" json:"userid" bson:"userid"`
	Version   Version `yaml:"version" json:"version" bson:"version"`
	VersionID string  `yaml:"versionId" json:"versionId" bson:"versionId"`
	ChangeLog string  `yaml:"changeLog" json:"changeLog" bson:"changeLog"`
	Config    Config  `yaml:"config" json:"config" bson:"config"`
	SortTime  int64   `yaml:"sortTime" json:"sortTime" bson:"sortTime"`
}

// Version 语义版本
type Version string

// ToString 将语义版本转换为字符串
func (v Version) ToString() string {
	return strings.TrimSpace(string(v))
}

// ParseVersion 解析语义版本
func (v Version) ParseVersion() (Semver, error) {
	v = v.TrimPrefix()

	matches := semverRegex.FindStringSubmatch(v.ToString())
	if matches == nil {
		return Semver{}, errors.New("invalid semantic version format")
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return Semver{}, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return Semver{}, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return Semver{}, fmt.Errorf("invalid patch version: %w", err)
	}
	return Semver{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// GetNextVersion 获取下一个语义版本
func (v Version) GetNextVersion() (string, error) {
	s, err := v.ParseVersion()
	if err != nil {
		return "", err
	}
	s.Patch++
	return fmt.Sprintf("v%d.%d.%d", s.Major, s.Minor, s.Patch), nil
}

// Compare 语义化版本比较
func (v Version) Compare(prev Version) (int, error) {
	v = v.TrimPrefix()
	prev = prev.TrimPrefix()

	newV, err := v.ParseVersion()
	if err != nil {
		return -1, err
	}

	oldV, err := prev.ParseVersion()
	if err != nil {
		return -1, err
	}

	return newV.Compare(oldV), nil
}

// TrimPrefix 去除前缀
func (v Version) TrimPrefix() Version {
	return Version(strings.TrimPrefix(v.ToString(), "v"))
}

// Semver 语义版本
type Semver struct {
	Major int `yaml:"major" json:"major" bson:"major"`
	Minor int `yaml:"minor" json:"minor" bson:"minor"`
	Patch int `yaml:"patch" json:"patch" bson:"patch"`
}

// Compare 两个语义版本的比较结果，返回：
//
//	-1 如果 s 小于 prev
//	1  如果 s 大于 prev
//	0  如果 s 等于 prev
func (s Semver) Compare(prev Semver) int {
	if s.Major > prev.Major {
		return 1
	}
	if s.Major < prev.Major {
		return -1
	}

	// Major 相同，比较 Minor 版本
	if s.Minor > prev.Minor {
		return 1
	}
	if s.Minor < prev.Minor {
		return -1
	}

	// Minor 相同，比较 Patch 版本
	if s.Patch > prev.Patch {
		return 1
	}
	if s.Patch < prev.Patch {
		return -1
	}

	// 所有版本号都相同
	return 0
}

// Config DAG配置
type Config string

// ParseToDag 反序列化dag配置
func (c Config) ParseToDag() (*Dag, error) {
	dag := &Dag{}
	err := json.Unmarshal([]byte(c), dag)
	if err != nil {
		return nil, err
	}

	if dag.ID == "" {
		return nil, errors.New("invalid dag config")
	}

	return dag, nil
}
