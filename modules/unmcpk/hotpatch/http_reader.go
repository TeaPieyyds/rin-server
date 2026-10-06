package hotpatch

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type PartialHttpReader struct {
	httpUrl    string
	httpClient *http.Client

	lastResponse *http.Response
	lastPosition int64

	seekPosition int64

	ContentSize int
}

func NewPartialHttpReader(httpUrl string, httpClient *http.Client) (*PartialHttpReader, error) {
	pr := &PartialHttpReader{
		httpUrl:    httpUrl,
		httpClient: httpClient,
	}

	resp, err := pr.requestWithMethod("HEAD", map[string]string{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server respond with code: %d", resp.StatusCode)
	}

	pr.ContentSize, err = strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (pr *PartialHttpReader) requestWithMethod(method string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, pr.httpUrl, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return pr.httpClient.Do(req)
}

func (pr *PartialHttpReader) ReadAt(p []byte, off int64) (int, error) {
	if pr.lastResponse != nil {
		if off == pr.lastPosition {
			n, err := pr.lastResponse.Body.Read(p)
			if err != nil {
				pr.lastResponse.Body.Close()
				pr.lastResponse = nil
			}
			pr.lastPosition += int64(n)
			return n, err
		} else {
			pr.lastResponse.Body.Close()
			pr.lastResponse = nil
		}
	}
	resp, err := pr.requestWithMethod("GET", map[string]string{
		"Range": fmt.Sprintf("bytes=%d-%d", off, pr.ContentSize),
	})
	if err != nil {
		return 0, err
	}

	n, err := resp.Body.Read(p)
	if err != nil {
		resp.Body.Close()
		return 0, err
	}
	pr.lastResponse = resp
	pr.lastPosition = off + int64(n)
	return n, err
}

func (pr *PartialHttpReader) Read(p []byte) (n int, err error) {
	n, err = pr.ReadAt(p, pr.seekPosition)
	pr.seekPosition += int64(n)
	return
}

func (pr *PartialHttpReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		pr.seekPosition = offset
	case io.SeekEnd:
		pr.seekPosition = int64(pr.ContentSize) - offset
	}
	return pr.seekPosition, nil
}

func (pr *PartialHttpReader) Close() error {
	if pr.lastResponse != nil {
		pr.lastResponse.Body.Close()
		pr.lastResponse = nil
	}
	return nil
}
