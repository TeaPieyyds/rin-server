package mcpk

import (
	"errors"
	"strings"
)

func (fs *MCPKFileSystem) Open(name string) ([]byte, error) {
	index := strings.LastIndexByte(name, '/')
	var directoryId uint32 = 0
	if index != -1 {
		dirName := name[:index]
		directoryId = StringID(dirName)
	}

	// load file entries inside directory entry
	fileEntries, ok := fs.FileEntries[directoryId]
	if !ok {
		if !fs.prepareFileEntries(directoryId) {
			return nil, errors.New("no directory found")
		}
		fileEntries = fs.FileEntries[directoryId]
	}

	fileId := StringID(name[index+1:])
	var file *FileEntry
	for _, e := range fileEntries {
		if e.Id == fileId {
			ele := e
			file = &ele
		}
	}

	if file == nil {
		return nil, errors.New("no file found")
	}

	data := fs.data[fs.offsetFileEntryEnd+file.Offset : fs.offsetFileEntryEnd+file.Offset+file.Size]

	return data, nil
}
