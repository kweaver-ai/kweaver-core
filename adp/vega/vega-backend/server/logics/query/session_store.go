// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package query 提供统一查询与 JOIN 能力
package query

import (
	"context"
	"sync"
	"time"

	"vega-backend/interfaces"
)

const (
	// SessionTTL session 生存时间，自最后一次访问起
	SessionTTL = 30 * time.Minute
	// CleanupInterval 定时清理间隔
	CleanupInterval = 5 * time.Minute
)

// memorySessionStore 内存实现的 QuerySessionStore
type memorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*querySession
	ttl      time.Duration
	stopCh   chan struct{}
}

type querySession struct {
	cursors    map[int]string // offset -> cursorEncoded
	lastAccess time.Time
}

// NewMemorySessionStore 创建内存 session 存储
func NewMemorySessionStore(ttl time.Duration) interfaces.QuerySessionStore {
	s := &memorySessionStore{
		sessions: make(map[string]*querySession),
		ttl:      ttl,
		stopCh:   make(chan struct{}),
	}
	if s.ttl == 0 {
		s.ttl = SessionTTL
	}
	go s.cleanupLoop()
	return s
}

// GetCursor 获取游标
func (s *memorySessionStore) GetCursor(ctx context.Context, queryID string, offset int) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, ok := s.sessions[queryID]
	if !ok || sess == nil {
		return "", false
	}
	cursor, ok := sess.cursors[offset]
	return cursor, ok
}

// SetCursor 设置游标
func (s *memorySessionStore) SetCursor(ctx context.Context, queryID string, offset int, cursorEncoded string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[queryID]
	if !ok {
		sess = &querySession{
			cursors:    make(map[int]string),
			lastAccess: time.Now(),
		}
		s.sessions[queryID] = sess
	}
	sess.cursors[offset] = cursorEncoded
	sess.lastAccess = time.Now()
	return nil
}

// Touch 刷新最后访问时间
func (s *memorySessionStore) Touch(ctx context.Context, queryID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[queryID]
	if ok && sess != nil {
		sess.lastAccess = time.Now()
	}
	return nil
}

// cleanupLoop 定时清理过期 session
func (s *memorySessionStore) cleanupLoop() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

func (s *memorySessionStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, sess := range s.sessions {
		if now.Sub(sess.lastAccess) > s.ttl {
			delete(s.sessions, id)
		}
	}
}
