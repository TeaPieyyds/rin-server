package hotpatch

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"github.com/Yeah114/unmcpk/mcpk"
)

func ExtractEngineMCPK(path string) error {
	v, err := FetchEngineVersion()
	if err != nil {
		return err
	}

	pr, err := NewPartialHttpReader(v.URL, httpClient)
	if err != nil {
		return err
	}
	defer pr.Close()
	zr, err := zip.NewReader(pr, int64(pr.ContentSize))
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "assets") && strings.HasSuffix(f.Name, "vanilla.mcp") {
			o, err := f.Open()
			if err != nil {
				return err
			}

			body, err := io.ReadAll(o)
			if err != nil {
				return err
			}

			mcpk.Unpack(body, path)
		}
	}

	return nil
}

type EngineVersion struct {
	URL     string `json:"url"`
	Version string `json:"version"`
}

func FetchEngineVersion() (*EngineVersion, error) {
	body, err := fetch("https://g79.update.netease.com/pack_list/production/g79_packlist_2")
	if err != nil {
		return nil, err
	}

	data := make(map[string]EngineVersion)
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	if v, ok := data["netease"]; ok {
		return &v, nil
	}

	return nil, errors.New("no package found")
}
