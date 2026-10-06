package hotpatch

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

type PatchlistResponse struct {
	URL     string   `json:"url"`
	NewURL  string   `json:"urlNew"`
	IOS     []string `json:"ios"`
	Android []string `json:"android"`
	Threads string   `json:"threadNum"`
}

func (p *PatchlistResponse) Latest() string {
	// compare with semver
	list := make([]string, len(p.Android))

	copy(list, p.Android)
	sort.Slice(list, func(i, j int) bool {
		vi := strings.Split(list[i], ".")
		vj := strings.Split(list[j], ".")

		for k := 0; k < len(vi) && k < len(vj); k++ {
			numVi, _ := strconv.Atoi(vi[k])
			numVj, _ := strconv.Atoi(vj[k])

			if numVi != numVj {
				return numVi < numVj
			}
		}

		return false
	})

	return list[len(list)-1]
}

func (p *PatchlistResponse) ListUniqueMCPK() ([]string, error) {
	hash := ""
	versions := make([]string, 0)

	for _, e := range p.Android {
		m, err := FetchManifest(p.NewURL, e)
		if err != nil {
			return nil, err
		}

		currentHash := m.GetHashMCPK()
		if currentHash != hash {
			versions = append(versions, e)
			hash = currentHash
		}
	}

	return versions, nil
}

func FetchPatchlist() (*PatchlistResponse, error) {
	body, err := fetch("https://g79.update.netease.com/patch_list/production/g79_rn_patchlist")
	if err != nil {
		return nil, err
	}

	var data PatchlistResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
