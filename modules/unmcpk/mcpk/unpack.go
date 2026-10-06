package mcpk

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Yeah114/unmcpk/pymarshal"
)

var rotor *RotorObj
var idRedirect = StringID("redirect.mcs")

func init() {
	asdf_dn := "j2h56ogodh3sk"
	asdf_dt := "=dziaq;"
	asdf_df := "|`o=5v7!\"-276"

	asdf_tm := strings.Repeat(asdf_dn, 4) + strings.Repeat(asdf_dt+asdf_dn+asdf_df, 5) + "$@" + strings.Repeat(asdf_dt, 7) + strings.Repeat(asdf_df, 2) + "&]`"

	rotor = NewRotorObj(6, asdf_tm)
}

func Unpack(dat []byte, path string) error {
	start := binary.LittleEndian.Uint32(dat[0x0c:])
	middle := binary.LittleEndian.Uint32(dat[0x10:])
	end := binary.LittleEndian.Uint32(dat[0x14:])
	//fmt.Printf("%x, %x, %x\n", start, middle, end)

	directory_table := dat[start:middle]
	var dirId, dirOff, dirSize, v1, v2, v3, _ uint32
	for i := 0; i < len(directory_table); i += 0xc {
		dirId = binary.LittleEndian.Uint32(directory_table[i:])
		dirOff = binary.LittleEndian.Uint32(directory_table[i+4:])
		dirSize = binary.LittleEndian.Uint32(directory_table[i+8:])
		index_table := dat[middle+dirOff : middle+dirOff+(dirSize*0x10)]
		for i := 0; i < len(index_table); i += 0x10 {
			v1 = binary.LittleEndian.Uint32(index_table[i:])
			v2 = binary.LittleEndian.Uint32(index_table[i+4:])
			v3 = binary.LittleEndian.Uint32(index_table[i+8:])
			_ = binary.LittleEndian.Uint32(index_table[i+12:])
			// if v1 == idRedirect {
			data, err := decrypt(dat[end+v2:end+v2+v3], v1)
			if err != nil {
				return fmt.Errorf("failed to decrypt file %x: %w", v1, err)
			}

			//println("size:", len(data))

			repaired, err := pymarshal.PycRepairer(data, v1 != idRedirect)
			if err != nil {
				return err
			}

			err = os.WriteFile(filepath.Join(path, fmt.Sprintf("%x_%x.pyc", dirId, v1)), repaired, 0777)
			if err != nil {
				return err
			}
			//fmt.Printf("%x\t%x\t%x\t%x\n", v1, v2, v3, v4)
			// }
		}
	}
	return nil
}

func decrypt(data []byte, fileId uint32) ([]byte, error) {
	decrypted := make([]byte, len(data))
	copy(decrypted, data)

	// decrypt
	if fileId == idRedirect {
		xorData := []byte("MCPK")
		for i := range xorData {
			decrypted[i] ^= xorData[i]
		}
	} else {
		rotor.DRegion(decrypted, true)
	}

	// unzip
	var unzipped bytes.Buffer
	reader, err := zlib.NewReader(bytes.NewBuffer(decrypted))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	_, err = io.Copy(&unzipped, reader)
	if err != nil && err != io.EOF {
		return nil, err
	}

	if fileId != idRedirect {
		decrypted = reverseData(unzipped.Bytes())
	} else {
		decrypted = unzipped.Bytes()
	}

	return decrypted, nil
}

func reverseData(data []byte) []byte {
	length := len(data)

	for i := 0; i < 130 && i < length; i++ {
		data[i] = data[i] ^ 156
	}

	for i := 0; i < length/2; i++ {
		data[i], data[length-i-1] = data[length-i-1], data[i]
	}

	return data
}
