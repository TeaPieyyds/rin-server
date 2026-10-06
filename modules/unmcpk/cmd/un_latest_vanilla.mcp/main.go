package main

import (
	"io"
	"os"
	"fmt"
	"time"
	"net/http"

	"github.com/Yeah114/unmcpk/mcpk"
	"github.com/Yeah114/unmcpk/hotpatch"
)

func main(){
	patchList, err := hotpatch.FetchPatchlist()
	if err != nil {
		panic(err)
	}

	manifest, err := hotpatch.FetchManifest(patchList.NewURL, patchList.Android[len(patchList.Android)-1])
	if err != nil {
		panic(err)
	}

	assetName := "vanilla.mcp"
	asset, ok := manifest.Assets[assetName]
	if !ok {
		panic(fmt.Errorf("%s does not exist", assetName))
	}

	url := manifest.BaseUrl + asset.PathUrl
	file, err := os.OpenFile(assetName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	var httpClient *http.Client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: nil,
		},
	}
	reader, err := hotpatch.NewPartialHttpReader(url, httpClient)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	
	_, err = file.Write(body)
	if err != nil {
		panic(err)
	}

	dir := "extracted"
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}

	err = mcpk.Unpack(body, dir)
	if err != nil {
		panic(err)
	}
}