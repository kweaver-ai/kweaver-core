// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package query

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/interfaces"
)

// StreamQuerySession 流式查询会话
type StreamQuerySession struct {
	QueryID      string
	Connector    string
	Database     string
	CatalogID    string
	Catalog      *interfaces.Catalog
	StreamSize   int // 流式查询每批数据量
	Offset       int // 当前查询的偏移量
	LastAccessed time.Time
	CreatedAt    time.Time
	OriginalSQL  string   // 原始SQL查询语句
	ResourceIds  []string // 资源ID列表
}

// StreamQueryManager 流式查询管理器
type StreamQueryManager struct {
	sessions map[string]*StreamQuerySession
	mu       sync.RWMutex
}

var (
	streamQueryManager     *StreamQueryManager
	streamQueryManagerOnce sync.Once
)

// GetStreamQueryManager 获取流式查询管理器单例
func GetStreamQueryManager() *StreamQueryManager {
	streamQueryManagerOnce.Do(func() {
		streamQueryManager = &StreamQueryManager{
			sessions: make(map[string]*StreamQuerySession),
		}
		// 启动清理goroutine
		go streamQueryManager.cleanupExpiredSessions()
	})
	return streamQueryManager
}

// CreateSession 创建新的流式查询会话
func (m *StreamQueryManager) CreateSession(connector, database, catalogID string, catalog *interfaces.Catalog, streamSize int, originalSQL string, resourceIds []string) (*StreamQuerySession, error) {
	queryID := uuid.New().String()

	// 如果streamSize未设置或小于等于0，使用默认值10000
	if streamSize <= 0 {
		streamSize = 10000
	}

	session := &StreamQuerySession{
		QueryID:      queryID,
		Connector:    connector,
		Database:     database,
		CatalogID:    catalogID,
		Catalog:      catalog,
		StreamSize:   streamSize,
		Offset:       0,
		LastAccessed: time.Now(),
		CreatedAt:    time.Now(),
		OriginalSQL:  originalSQL,
		ResourceIds:  resourceIds,
	}

	m.mu.Lock()
	m.sessions[queryID] = session
	m.mu.Unlock()

	logger.Infof("Created stream query session: %s for connector: %s, database: %s, stream_size: %d", queryID, connector, database, streamSize)
	return session, nil
}

// GetSession 获取流式查询会话
func (m *StreamQueryManager) GetSession(queryID string) (*StreamQuerySession, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[queryID]
	if ok {
		session.LastAccessed = time.Now()
	}
	return session, ok
}

// RemoveSession 移除流式查询会话
func (m *StreamQueryManager) RemoveSession(queryID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[queryID]; ok {
		delete(m.sessions, queryID)
		logger.Infof("Removed stream query session: %s", queryID)
	}
}

// cleanupExpiredSessions 清理过期的会话
func (m *StreamQueryManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for queryID, session := range m.sessions {
			// 清理超过30分钟未使用的会话
			if now.Sub(session.LastAccessed) > 30*time.Minute {
				delete(m.sessions, queryID)
				logger.Infof("Cleaned up expired stream query session: %s", queryID)
			}
		}
		m.mu.Unlock()
	}
}
