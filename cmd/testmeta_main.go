package main

import (
"fmt"
"github.com/Yeah114/g79client"
)

func main() {
m, err := g79client.GetGlobalG79PatchMetadata()
if err != nil {
fmt.Println("ERR:", err)
return
}
fmt.Printf("Version=%s\nResourcesHash=%s\n", m.Version, m.ResourcesHash)
}
