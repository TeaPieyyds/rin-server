package mcpk

import (
	"encoding/binary"
)

func (fs *MCPKFileSystem) prepareFileEntries(id uint32) bool {
	var directoryEntryTable *DirectoryEntry

	for _, e := range fs.DirectoryEntryList {
		if e.Id == id {
			ele := e
			directoryEntryTable = &ele
		}
	}

	if directoryEntryTable == nil {
		return false
	}

	fileEntries := make([]FileEntry, directoryEntryTable.Size)
	data := fs.data[fs.offsetFileEntry+directoryEntryTable.Offset : fs.offsetFileEntryEnd]
	for i := uint32(0); i < directoryEntryTable.Size; i++ {
		baseOffset := i * 0x10
		fileEntries[i] = FileEntry{
			Id:      binary.LittleEndian.Uint32(data[baseOffset:]),
			Offset:  binary.LittleEndian.Uint32(data[baseOffset+4:]),
			Size:    binary.LittleEndian.Uint32(data[baseOffset+8:]),
			Unknown: binary.LittleEndian.Uint32(data[baseOffset+12:]),
		}
	}

	fs.FileEntries[id] = fileEntries

	return true
}
