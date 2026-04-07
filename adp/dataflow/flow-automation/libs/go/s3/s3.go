package s3

import (
	"os"
	"sync"

	"github.com/goccy/go-yaml"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
)

const (
	S3ConfigPath = "/conf/s3.yaml"
)

var (
	s3Ins  *S3
	s3Once sync.Once
)

type S3 struct {
	connMap map[string]*S3Connection `json:"-"`

	Default     string          `json:"default"`
	Connections []*S3Connection `json:"connections"`
}

func NewS3() *S3 {
	s3Once.Do(func() {
		s3Ins = &S3{
			connMap: make(map[string]*S3Connection),
		}
		f, err := os.Open(S3ConfigPath)
		if err != nil {
			commonLog.NewLogger().Errorf("open s3.yaml failed: %s", err.Error())
			return
		}
		defer f.Close()
		decoder := yaml.NewDecoder(f)
		if err := decoder.Decode(s3Ins); err != nil {
			commonLog.NewLogger().Errorf("decode s3.yaml failed: %s", err.Error())
			return
		}
		for _, conn := range s3Ins.Connections {
			if err := conn.InitClient(); err != nil {
				commonLog.NewLogger().Errorf("init s3 connection %s failed: %s", conn.Name, err.Error())
				continue
			}
			s3Ins.connMap[conn.Name] = conn
		}
	})
	return s3Ins
}

func (s *S3) GetConnection(name string) *S3Connection {
	conn, ok := s.connMap[name]
	if !ok {
		return nil
	}
	return conn
}

func (s *S3) GetDefaultConnection() *S3Connection {
	return s.GetConnection(s.Default)
}

func (s *S3) GetAvailableConnections() []*S3Connection {
	var availableConns []*S3Connection
	for _, conn := range s.Connections {
		if conn.client != nil {
			availableConns = append(availableConns, conn)
		}
	}
	return availableConns
}

func (s *S3) GetAvailableConnection() *S3Connection {
	defaultConn := s.GetDefaultConnection()
	if defaultConn != nil && defaultConn.client != nil {
		return defaultConn
	}

	for _, conn := range s.Connections {
		if conn.client != nil {
			return conn
		}
	}
	return nil
}
