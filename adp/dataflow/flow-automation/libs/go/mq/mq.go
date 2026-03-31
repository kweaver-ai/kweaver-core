package mq

import (
	"os"

	msqclient "github.com/kweaver-ai/proton-mq-sdk-go"
	"gopkg.in/yaml.v2"
)

//go:generate mockgen -package mock_drivenadapters -source ../mq/mq.go -destination ../mock/mock_drivenadapters/mq_mock.go

var mqClientSingleton *mqClient

type mqClient struct {
	mqClient                 msqclient.ProtonMQClient
	pollIntervalMilliseconds int64
	maxInFlight              int
}

var _ MQClient = &mqClient{}

// MQClient client methods
type MQClient interface {
	// Subscribe subscribe
	Subscribe(topic, channel string, handler func([]byte) error) error
	// Publish publish
	Publish(topic string, message []byte) (err error)
	// Close close
	Close()
}

// Publish pub msg
func (m *mqClient) Publish(topic string, message []byte) (err error) {
	return m.mqClient.Pub(topic, message)
}

// Subscribe sub msg
func (m *mqClient) Subscribe(topic, channel string, handler func([]byte) error) error {
	return m.mqClient.Sub(topic, channel, handler, m.pollIntervalMilliseconds, m.maxInFlight)
}

// Close  close mq client
func (m *mqClient) Close() {
	m.mqClient.Close()
}

type MQConfig struct {
	ConfigPath               string
	PollIntervalMilliseconds int64
	MaxInFlight              int
	msqclient.ProtonMQInfo
}

// InitMQClient init MQ Client
func InitMQClient(config *MQConfig) error {
	var err error
	var mq msqclient.ProtonMQClient
	mqConfigPath := extractMQInfoFromServiceAccess(config.ConfigPath)
	if mqConfigPath == "" {
		// Only add authentication options if mechanism is provided
		var opts []msqclient.ClientOpt
		if config.Auth.Mechanism != "" {
			opts = []msqclient.ClientOpt{
				msqclient.AuthMechanism(config.Auth.Mechanism),
				msqclient.UserInfo(config.Auth.Username, config.Auth.Password),
			}
		}
		mq, err = msqclient.NewProtonMQClient(config.Host, config.Port, config.LookupdHost, config.LookupdPort, config.MQType, opts...)
		if err != nil {
			return err
		}
	} else {
		mq, err = msqclient.NewProtonMQClientFromFile(mqConfigPath)
		if err != nil {
			return err
		}
	}

	mqClientSingleton = &mqClient{
		mqClient:                 mq,
		pollIntervalMilliseconds: config.PollIntervalMilliseconds,
		maxInFlight:              config.MaxInFlight,
	}

	return nil
}

// NewMQClient new
func NewMQClient() MQClient {
	return mqClientSingleton
}

func extractMQInfoFromServiceAccess(configPath string) string {
	if configPath == "" {
		return ""
	}
	const mqConfigPath string = "/tmp/service-access-mq.yaml"
	yamlIn, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}
	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(yamlIn, &m)
	if err != nil {
		return ""
	}
	_, ok := m["mqType"]
	if ok {
		return configPath
	}
	mq, ok := m["mq"].(map[interface{}]interface{})
	if !ok {
		return ""
	}
	yamlOut, err := yaml.Marshal(mq)
	if err != nil {
		return ""
	}
	err = os.WriteFile(mqConfigPath, yamlOut, 0744)
	if err != nil {
		return ""
	}
	return mqConfigPath
}
