package opensearchcmp

import (
	"crypto/tls"
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/opensearch-project/opensearch-go"
)

func (o *OpsCmp) newClient() (err error) {
	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{
			o.address,
		},
		Username: o.username,
		Password: o.password,
	})

	if cenvhelper.IsLocalDev() {
		client, err = opensearch.NewClient(opensearch.Config{
			Addresses: []string{
				o.address,
			},
			Username: o.username,
			Password: o.password,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			},
		})
	}

	if err != nil {
		return
	}

	o.client = client

	return
}
