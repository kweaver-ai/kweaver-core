package sandbox

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	lock "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/lock"
	libstore "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/store"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

const (
	SandboxLockPrefix       = "automation:sandbox_session"
	SandboxSessionKeyPrefix = "automation:sandbox_session:id:"
	SandboxLockTTL          = 30 * time.Second
	SandboxLockWaitTTL      = 60 * time.Second
	DefaultIntervalSec      = 1
	SessionPollInterval     = 3 * time.Second
	SessionPollTimeout      = 30 * time.Second
)

func MergeSessionConfig(nodeConfig *drivenadapters.SandboxSessionConfig, globalConfig common.Sandbox) *drivenadapters.SandboxSessionConfig {
	config := &drivenadapters.SandboxSessionConfig{}

	config.TemplateID = globalConfig.TemplateID
	config.CPU = globalConfig.Cpu
	config.Memory = globalConfig.Memory
	config.Disk = globalConfig.Disk
	config.Timeout = globalConfig.Timeout

	if nodeConfig != nil {
		if nodeConfig.TemplateID != "" {
			config.TemplateID = nodeConfig.TemplateID
		}
		if nodeConfig.CPU != "" {
			config.CPU = nodeConfig.CPU
		}
		if nodeConfig.Memory != "" {
			config.Memory = nodeConfig.Memory
		}
		if nodeConfig.Disk != "" {
			config.Disk = nodeConfig.Disk
		}
		if nodeConfig.Timeout > 0 {
			config.Timeout = nodeConfig.Timeout
		}
		if len(nodeConfig.Dependencies) > 0 {
			config.Dependencies = nodeConfig.Dependencies
		}
	}

	if config.Timeout <= 0 {
		config.Timeout = 30
	}

	return config
}

func waitForSessionRunning(ctx context.Context, sessionID string) error {
	log := traceLog.WithContext(ctx)
	sandboxClient := drivenadapters.NewSandbox()

	pollCtx, cancel := context.WithTimeout(ctx, SessionPollTimeout)
	defer cancel()

	ticker := time.NewTicker(SessionPollInterval)
	defer ticker.Stop()

	log.Infof("[Sandbox] Waiting for session %s to be running...", sessionID)

	for {
		select {
		case <-pollCtx.Done():
			return fmt.Errorf("session %s timeout waiting for running status", sessionID)
		case <-ticker.C:
			session, err := sandboxClient.GetSession(ctx, sessionID)
			if err != nil {
				log.Warnf("[Sandbox] Get session %s failed: %v", sessionID, err)
				continue
			}

			switch session.Status {
			case drivenadapters.SessionStatusRunning:
				log.Infof("[Sandbox] Session %s is now running", sessionID)
				return nil
			case drivenadapters.SessionStatusCreating:
				log.Infof("[Sandbox] Session %s is still creating, waiting...", sessionID)
			default:
				log.Warnf("[Sandbox] Session %s is unavailable with status: %s", sessionID, session.Status)
				return fmt.Errorf("session %s is unavailable with status: %s", sessionID, session.Status)
			}
		}
	}
}

func checkAndReuseSession(ctx context.Context, sessionID string) bool {
	log := traceLog.WithContext(ctx)
	sandboxClient := drivenadapters.NewSandbox()

	session, serr := sandboxClient.GetSession(ctx, sessionID)
	if serr != nil {
		log.Warnf("[Sandbox] Get session %s failed: %v", sessionID, serr)
		return false
	}

	if session == nil {
		return false
	}

	switch session.Status {
	case drivenadapters.SessionStatusRunning:
		log.Infof("[Sandbox] Reuse existing running session: %s", sessionID)
		return true
	case drivenadapters.SessionStatusCreating:
		log.Infof("[Sandbox] Session %s is creating, waiting for it to be running...", sessionID)
		serr := waitForSessionRunning(ctx, sessionID)
		if serr != nil {
			log.Warnf("[Sandbox] Session %s failed to become running: %v", sessionID, serr)
			return false
		}
		log.Infof("[Sandbox] Session %s is now running, reusing it", sessionID)
		return true
	case drivenadapters.SessionStatusFailed, drivenadapters.SessionStatusCompleted:
		log.Infof("[Sandbox] Session %s is invalid (status: %s), creating new one", sessionID, session.Status)
		return false
	default:
		log.Warnf("[Sandbox] Session %s is unavailable with status: %s, creating new one", sessionID, session.Status)
		return false
	}
}

func GetOrCreateSession(ctx context.Context, config *drivenadapters.SandboxSessionConfig) (string, error) {
	log := traceLog.WithContext(ctx)
	redisClient := libstore.NewRedis().GetClient()

	configBytes, _ := json.Marshal(config)
	configHash := Hash(string(configBytes))

	sessionKey := SandboxSessionKeyPrefix + configHash
	lockKey := SandboxLockPrefix + ":" + configHash

	sandboxClient := drivenadapters.NewSandbox()

	sessionID, err := redisClient.Get(ctx, sessionKey).Result()
	if err == nil && sessionID != "" {
		reused := checkAndReuseSession(ctx, sessionID)
		if reused {
			return sessionID, nil
		}
	}

	lockClient := lock.NewDistributeLock(libstore.NewRedis(), lockKey, "sandbox_executor")
	lockCtx, cancel := context.WithTimeout(ctx, SandboxLockWaitTTL)
	defer cancel()

	err = lockClient.TryLock(lockCtx, SandboxLockTTL, false)
	if err != nil {
		log.Warnf("[Sandbox] Acquire lock failed: %v, trying to reuse session created by other goroutine", err)
		sessionID, err = redisClient.Get(ctx, sessionKey).Result()
		if err == nil && sessionID != "" {
			reused := checkAndReuseSession(ctx, sessionID)
			if reused {
				return sessionID, nil
			}
		}
		return "", fmt.Errorf("acquire lock failed: %w", err)
	}
	defer lockClient.Release()

	sessionID, err = redisClient.Get(ctx, sessionKey).Result()
	if err == nil && sessionID != "" {
		reused := checkAndReuseSession(ctx, sessionID)
		if reused {
			return sessionID, nil
		}
	}

	log.Infof("[Sandbox] Creating new session for config hash: %s", configHash)

	var deps []drivenadapters.SessionDependency
	for _, dep := range config.Dependencies {
		deps = append(deps, dep)
	}

	createReq := &drivenadapters.SandboxSessionConfig{
		TemplateID:   config.TemplateID,
		CPU:          config.CPU,
		Memory:       config.Memory,
		Disk:         config.Disk,
		Timeout:      config.Timeout,
		Dependencies: deps,
	}

	newSession, err := sandboxClient.CreateSession(ctx, createReq)
	if err != nil {
		return "", fmt.Errorf("create session failed: %w", err)
	}

	log.Infof("[Sandbox] New session created: %s, checking status...", newSession.ID)

	session, err := sandboxClient.GetSession(ctx, newSession.ID)
	if err != nil {
		return "", fmt.Errorf("get new session %s failed: %w", newSession.ID, err)
	}

	switch session.Status {
	case drivenadapters.SessionStatusRunning:
		log.Infof("[Sandbox] New session %s is already running", newSession.ID)
	case drivenadapters.SessionStatusCreating:
		log.Infof("[Sandbox] New session %s is creating, waiting for it to be running...", newSession.ID)
		err := waitForSessionRunning(ctx, newSession.ID)
		if err != nil {
			return "", fmt.Errorf("new session %s failed to become running: %w", newSession.ID, err)
		}
		log.Infof("[Sandbox] New session %s is now running", newSession.ID)
	case drivenadapters.SessionStatusFailed:
		return "", fmt.Errorf("new session %s failed during initialization", newSession.ID)
	case drivenadapters.SessionStatusCompleted:
		return "", fmt.Errorf("new session %s completed unexpectedly", newSession.ID)
	default:
		return "", fmt.Errorf("new session %s has unknown status: %s", newSession.ID, session.Status)
	}

	err = redisClient.Set(ctx, sessionKey, newSession.ID, 24*time.Hour).Err()
	if err != nil {
		log.Warnf("[Sandbox] Save session key failed: %v", err)
	}

	return newSession.ID, nil
}

func Hash(s string) string {
	data := []byte(s)
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}
