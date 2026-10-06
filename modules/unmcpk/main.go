package unmcpk

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"io"
	"os/exec"
	"fmt"
	"strings"

	"github.com/Yeah114/unmcpk/mcpk"
	"github.com/Yeah114/unmcpk/pymarshal"
)

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

func DecryptDynamicMCP(decrypted []byte) ([]byte, error) {

	asdf_dn := "j2h56ogodh3sk"
	asdf_dt := "=dziaq;"
	asdf_df := "|`o=5v7!\"-276"

	asdf_tm := strings.Repeat(asdf_dn, 4) + strings.Repeat(asdf_dt+asdf_dn+asdf_df, 5) + "$@" + strings.Repeat(asdf_dt, 7) + strings.Repeat(asdf_df, 2) + "&]`"

	rotor := mcpk.NewRotorObj(6, asdf_tm)

	rotor.DRegion(decrypted, true)

	var unzipped bytes.Buffer
	reader, err := zlib.NewReader(bytes.NewBuffer(decrypted))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(&unzipped, reader)
	if err != io.EOF && err != nil {
		return nil, err
	}
	reader.Close()

	decrypted = reverseData(unzipped.Bytes())

	repaired, err := pymarshal.PycRepairer(decrypted, true)
	if err != nil {
		return nil, err
	}
	return repaired, nil

}

func EncryptDynamicMCP(pyc []byte) ([]byte, error) {
	asdfDn := "j2h56ogodh3sk"
	asdfDt := "=dziaq;"
	asdfDf := "|`o=5v7!\"-276"

	asdfTm := strings.Repeat(asdfDn, 4) + strings.Repeat(asdfDt+asdfDn+asdfDf, 5) + "$@" + strings.Repeat(asdfDt, 7) + strings.Repeat(asdfDf, 2) + "&]`"

	marshalEncrypted, err := pymarshal.PycEncryptor(pyc, pymarshal.EncryptOptions{
		StringVersion: 2,
		CodeVersion:   1,
		OpcodeType:    0,
	})
	if err != nil {
		return nil, fmt.Errorf("PycEncryptor 错误: %v", err)
	}

	payload := make([]byte, len(marshalEncrypted))
	copy(payload, marshalEncrypted)
	payload = reverseData(payload)

	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write(payload); err != nil {
		zw.Close()
		return nil, fmt.Errorf("zlib 写入错误: %v", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("zlib 关闭错误: %v", err)
	}

	out := make([]byte, compressed.Len())
	copy(out, compressed.Bytes())

	rotor := mcpk.NewRotorObj(6, asdfTm)
	rotor.ERegion(out, true)
	return out, nil
}

func CompileDynamicMCP(pySource, filename, python2Path string) ([]byte, error) {
	const pythonScript = `import binascii
import marshal
import struct
import sys

source = binascii.unhexlify(sys.argv[1])
filename = sys.argv[2]

code_obj = compile(source, filename, 'exec')
magic = '\x03\xf3\r\n'
timestamp = struct.pack('<I', 0)
bytecode = marshal.dumps(code_obj)

sys.stdout.write(binascii.hexlify(magic + timestamp + bytecode))
`
	inputHex := hex.EncodeToString([]byte(pySource))
	cmd := exec.Command(python2Path, "-c", pythonScript, inputHex, filename)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("命令执行错误: %v, 输出: %s", err, string(out))
	}

	decoded, err := hex.DecodeString(strings.TrimSpace(string(out)))
	if err != nil {
		return nil, fmt.Errorf("解码错误: %v", err)
	}
	fmt.Println(decoded)

	encrypted, err := EncryptDynamicMCP(decoded)
	if err != nil {
		return nil, fmt.Errorf("加密错误: %v", err)
	}

	return encrypted, nil
}

func DecompileDynamicMCP(decrypted []byte, python3Path string) (string, error) {
	repaired, err := DecryptDynamicMCP(decrypted)
	if err != nil {
		return "", err
	}
	const pythonScript = `import binascii
import io
import os
import sys
import uncompyle6
def decompile_pyc_hex(pyc_hex_str):
 pyc_file_data=binascii.unhexlify(pyc_hex_str)
 temp_pyc='temp.pyc'
 with open(temp_pyc,'wb+') as f:
  f.write(pyc_file_data)
 output_stream=io.StringIO()
 uncompyle6.decompile_file(temp_pyc, output_stream)
 decompiled_source=output_stream.getvalue()
 output_stream.close()
 os.remove(temp_pyc)
 return binascii.hexlify(decompiled_source.encode("utf-8")).decode("utf-8")
if len(sys.argv) != 2:
 sys.exit(1)
input_pyc_hex=sys.argv[1]
result_hex=decompile_pyc_hex(input_pyc_hex)
print(result_hex)`
	inputHex := hex.EncodeToString(repaired)
	cmd := exec.Command(python3Path, "-c", pythonScript, inputHex)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	decoded, err := hex.DecodeString(strings.TrimSpace(string(out)))
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}