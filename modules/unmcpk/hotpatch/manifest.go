package hotpatch

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
)

type Manifest struct {
	BaseUrl string `json:""`

	Assets map[string]struct {
		MD5      string `json:"md5"`
		MD5Check string `json:"md5check"`
		PathUrl  string `json:"pathUrl"`
		Size     uint32 `json:"size"`
	} `json:"assets"`
	EngineVersion string `json:"engineVersion"`
	Version       string `json:"version"`
}

func FetchManifest(url, version string) (*Manifest, error) {
	baseUrl := url + "/android_" + version + "/" + version + "/android/"

	body, err := fetch(baseUrl + "manifest.zip")
	if err != nil {
		return nil, err
	}

	r, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, err
	}
	f := r.File[0]

	file, err := f.Open()
	if err != nil {
		return nil, err
	}
	body, err = io.ReadAll(file)
	file.Close()
	if err != nil {
		return nil, err
	}

	var data Manifest
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	data.BaseUrl = baseUrl

	return &data, nil
}
