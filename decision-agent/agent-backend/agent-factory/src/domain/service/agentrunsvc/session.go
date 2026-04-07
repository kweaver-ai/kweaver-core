package agentsvc

import (
	"sync"

	agentresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/resp"
)

type Session struct {
	sync.RWMutex   // 添加读写锁
	ConversationID string
	TempMsgResp    agentresp.ChatResp
	Signal         chan struct{}
	IsResuming     bool
}

// 添加 Session 的方法
func (s *Session) UpdateTempMsgResp(resp agentresp.ChatResp) {
	s.Lock()
	defer s.Unlock()
	s.TempMsgResp = resp
}

func (s *Session) GetTempMsgResp() agentresp.ChatResp {
	s.RLock()
	defer s.RUnlock()

	return s.TempMsgResp
}

func (s *Session) GetSignal() chan struct{} {
	s.RLock()
	defer s.RUnlock()

	return s.Signal
}

func (s *Session) GetIsResuming() bool {
	s.RLock()
	defer s.RUnlock()

	return s.IsResuming
}

func (s *Session) SetIsResuming(isResuming bool) {
	s.Lock()
	defer s.Unlock()
	s.IsResuming = isResuming
}

func (s *Session) SetSignal(signal chan struct{}) {
	s.Lock()
	defer s.Unlock()
	s.Signal = signal
}

func (s *Session) CloseSignal() {
	s.Lock()
	defer s.Unlock()

	if s.Signal != nil {
		close(s.Signal)
		// NOTE: 关闭后，将信号设置为nil
		s.Signal = nil
	}
}

func (s *Session) SendSignal() {
	s.RLock()
	defer s.RUnlock()

	if s.Signal != nil && s.IsResuming {
		s.Signal <- struct{}{}
	}
}
