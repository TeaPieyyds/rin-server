package hotpatch

import (
	"errors"
	"github.com/Yeah114/unmcpk/mcpk"
)

func (m *Manifest) ExtractAllMCPK(strictResolve bool, path string) error {
	list := []string{"vanilla.mcp", "vanilla_patch.mcp"}

	for _, e := range list {
		bl, err := m.ExtractMCPK(e, path)
		if err != nil {
			return err
		} else if !bl && strictResolve {
			return errors.New("no file found as " + e)
		}
	}

	return nil
}

func (m *Manifest) GetHashMCPK() string {
	list := []string{"vanilla.mcp", "vanilla_patch.mcp"}
	hash := ""

	for _, e := range list {
		if a, ok := m.Assets[e]; ok {
			hash += a.MD5
		}
	}

	return hash
}

func (m *Manifest) ExtractMCPK(filename string, path string) (bool, error) {
	if _, ok := m.Assets[filename]; ok {
		body, err := fetch(m.BaseUrl + filename)
		if err != nil {
			return false, err
		}

		mcpk.Unpack(body, path)

		return true, nil
	} else {
		return false, nil
	}
}
