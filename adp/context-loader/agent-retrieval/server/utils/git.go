// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	commitTextTag int = iota
	commitTextHash
	commitTextTime
	commitTextName
	commitTextMessage

	defaultCommitTag = "0.0.0"
	// CommitHashNull 无commit hash
	CommitHashNull = "null"
	// CommitHashLen Commit hash 长度
	CommitHashLen = 32
)

// git describe --tags --abbrev=0 > commit.txt
// git show -s --format=%H%n%ci%n%cn%n%s >> commit.txt

// LoadGitCommitInfoFromFile load
func LoadGitCommitInfoFromFile(path string) (info GitCommitInfo, err error) {
	// load config file
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		info.Tag = defaultCommitTag
		info.Hash = "null"
		return
	}
	arr := strings.Split(string(data), "\n")
	for i, v := range arr {
		v = strings.TrimSpace(v)
		switch i {
		case commitTextTag:
			if v == "" {
				v = defaultCommitTag
			}
			info.Tag = v
		case commitTextHash:
			if len(v) < CommitHashLen {
				info.Hash = CommitHashNull
			} else {
				info.Hash = v
			}
		case commitTextTime:
			t, err := time.Parse("2006-01-02 15:04:05 -0700", v)
			if err == nil {
				info.Time = t
			}
		case commitTextName:
			info.Name = v
		case commitTextMessage:
			info.Message = v
		default:
			return
		}
	}
	return
}

// GitCommitInfo 提交信息
type GitCommitInfo struct {
	Hash    string
	Time    time.Time
	Name    string
	Message string
	Tag     string
}

// Version 获取version
func (info *GitCommitInfo) Version(hl int) string {
	hash := info.Hash
	if len(info.Hash) >= hl {
		hash = info.Hash[:hl]
	}
	return fmt.Sprintf("%s-%s-%s", info.Tag, info.Time.Format("20060102150405"), hash)
}
