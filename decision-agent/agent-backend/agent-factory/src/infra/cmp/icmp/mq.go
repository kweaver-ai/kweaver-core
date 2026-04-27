package icmp

type IMQClient interface {
	Publish(string, []byte) (err error)
	Subscribe(topic, channel string, cmd func([]byte) error) (err error)
	Close()
}
