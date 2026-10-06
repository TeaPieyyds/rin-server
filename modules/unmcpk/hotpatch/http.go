package hotpatch

import (
	"errors"
	"io"
	"net/http"
	"time"
)

var httpClient *http.Client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		Proxy: nil,
	},
}

func fetch(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("status: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}
