package pymarshal

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func le32(n int32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], uint32(n))
	return b[:]
}

func TestEncodeOpcodeBytesRoundtripViaRepairOpcode(t *testing.T) {
	pyBytecode := []byte{
		opcodes["LOAD_CONST"], 0, 0,
		opcodes["RETURN_VALUE"],
	}

	for opcodeType := range neteaseOpcodes {
		s := &encryptState{
			opts:           EncryptOptions{StringVersion: 2, CodeVersion: 1, OpcodeType: opcodeType},
			opcodeNameByPy: invertOpcodeMap(opcodes),
		}

		netBytes, err := s.encodeOpcodeBytes(pyBytecode, opcodeType)
		if err != nil {
			t.Fatalf("encodeOpcodeBytes(opcodeType=%x): %v", opcodeType, err)
		}

		in := append([]byte{TYPE_STRING}, append(le32(int32(len(netBytes))), netBytes...)...)
		var out bytes.Buffer
		if err := repairOpcode(bytes.NewReader(in), &out, opcodeType); err != nil {
			t.Fatalf("repairOpcode(opcodeType=%x): %v", opcodeType, err)
		}

		want := append([]byte{TYPE_STRING}, append(le32(int32(len(pyBytecode))), pyBytecode...)...)
		if !bytes.Equal(out.Bytes(), want) {
			t.Fatalf("roundtrip mismatch for opcodeType=%x", opcodeType)
		}
	}
}

func TestMarshalEncryptStringRoundtripViaRepairMarshal(t *testing.T) {
	plain := append([]byte{TYPE_STRING}, append(le32(5), []byte("hello")...)...)

	enc, err := MarshalEncrypt(plain, EncryptOptions{StringVersion: 2, CodeVersion: 1, OpcodeType: 0})
	if err != nil {
		t.Fatal(err)
	}

	depth = 0
	strings = make([][]byte, 0)

	var out bytes.Buffer
	if err := RepairMarshal(bytes.NewReader(enc), &out); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), plain) {
		t.Fatal("string roundtrip mismatch")
	}
}

func TestPycEncryptorSimpleCodeObjectDoesNotEOF(t *testing.T) {
	pyc := []byte{
		0x03, 0xF3, 0x0D, 0x0A, 0x00, 0x00, 0x00, 0x00,
		0x63,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00,
		0x40, 0x00, 0x00, 0x00,
		0x73, 0x05, 0x00, 0x00, 0x00, 0x48, 0x64, 0x00, 0x00, 0x53,
		0x28, 0x01, 0x00, 0x00, 0x00, 0x4E,
		0x28, 0x00, 0x00, 0x00, 0x00,
		0x28, 0x00, 0x00, 0x00, 0x00,
		0x28, 0x00, 0x00, 0x00, 0x00,
		0x28, 0x00, 0x00, 0x00, 0x00,
		0x73, 0x08, 0x00, 0x00, 0x00, 0x3C, 0x73, 0x74, 0x72, 0x69, 0x6E, 0x67, 0x3E,
		0x74, 0x08, 0x00, 0x00, 0x00, 0x3C, 0x6D, 0x6F, 0x64, 0x75, 0x6C, 0x65, 0x3E,
		0x01, 0x00, 0x00, 0x00,
		0x74, 0x00, 0x00, 0x00, 0x00,
	}

	_, err := PycEncryptor(pyc, EncryptOptions{StringVersion: 2, CodeVersion: 1, OpcodeType: 0})
	if err != nil {
		t.Fatalf("PycEncryptor returned error: %v", err)
	}
}
