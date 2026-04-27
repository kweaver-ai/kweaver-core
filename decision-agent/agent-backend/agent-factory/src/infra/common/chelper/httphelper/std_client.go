package httphelper

import (
	"net/http"
	"time"
)

func GetClient(timeout time.Duration) (client *http.Client) {
	tran := GetDefaultTp()

	client = &http.Client{
		Transport: tran,
		Timeout:   timeout,
	}

	return
}

func GetDefaultStdClient() *http.Client {
	return defaultStdClient
}
