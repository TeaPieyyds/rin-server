package mcpk

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// I don't want to mess its name with ModCoderPack
type MCPKFileSystem struct {
	data []byte

	offsetDirectoryEntry uint32
	offsetFileEntry      uint32
	offsetFileEntryEnd   uint32

	DirectoryEntryList []DirectoryEntry
	FileEntries        map[uint32][]FileEntry
}

type DirectoryEntry struct {
	Id     uint32
	Offset uint32
	Size   uint32
}

type FileEntry struct {
	Id      uint32
	Offset  uint32
	Size    uint32
	Unknown uint32
}

// replicates _ZN13MCPFileSystem5_InitEPKc
func NewMCPKFileSystem(data []byte) (*MCPKFileSystem, error) {
	if !bytes.Equal(data[0:4], []byte("MCPK")) {
		return nil, errors.New("invalid file format")
	}

	fs := &MCPKFileSystem{
		data:                 data,
		offsetDirectoryEntry: binary.LittleEndian.Uint32(data[0x0c:]),
		offsetFileEntry:      binary.LittleEndian.Uint32(data[0x10:]),
		offsetFileEntryEnd:   binary.LittleEndian.Uint32(data[0x14:]),
		FileEntries:          make(map[uint32][]FileEntry),
	}

	// read directory entry
	directoryEntryCount := (fs.offsetFileEntry - fs.offsetDirectoryEntry) / 0xC
	if directoryEntryCount == 0 {
		return nil, errors.New("invalid directory table")
	}

	fs.DirectoryEntryList = make([]DirectoryEntry, directoryEntryCount)
	directory_table := data[fs.offsetDirectoryEntry:fs.offsetFileEntry]
	var v1, v2, v3 uint32
	for i := 0; i < len(directory_table); i += 0xc {
		v1 = binary.LittleEndian.Uint32(directory_table[i:])
		v2 = binary.LittleEndian.Uint32(directory_table[i+4:])
		v3 = binary.LittleEndian.Uint32(directory_table[i+8:])
		fs.DirectoryEntryList[i/0xc] = DirectoryEntry{
			Id:     v1,
			Offset: v2,
			Size:   v3,
		}
	}

	return fs, nil
}
