package pymarshal

import (
	"bytes"
	"crypto/rc4"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type EncryptOptions struct {
	// StringVersion selects which NetEase string codec to use:
	// 0 => XOR 0x8d, 1 => RC4(0xA7,0x0D,0x37,0x7A), 2 => RC4(long key).
	StringVersion int

	// CodeVersion selects which NetEase code object layout to emit:
	// 0 => TYPE_NETEASE_CODE_0, 1 => TYPE_NETEASE_CODE_1, 2 => TYPE_NETEASE_CODE_2.
	CodeVersion int

	// OpcodeType selects the NetEase opcode table ID to use when scrambling bytecode.
	OpcodeType uint32
}

var (
	ErrUnsupportedStringVersion = errors.New("unsupported netease string version")
	ErrUnsupportedCodeVersion   = errors.New("unsupported netease code version")
)

func (o EncryptOptions) normalize() (EncryptOptions, error) {
	if o.StringVersion < 0 || o.StringVersion > 2 {
		return EncryptOptions{}, ErrUnsupportedStringVersion
	}
	if o.CodeVersion < 0 || o.CodeVersion > 2 {
		return EncryptOptions{}, ErrUnsupportedCodeVersion
	}
	return o, nil
}

// PycEncryptor converts a standard Python 2.x marshal payload (optionally with a
// `PycHeader` prefix) into the NetEase-encrypted marshal format understood by
// `PycRepairer`.
func PycEncryptor(pycOrMarshal []byte, opts EncryptOptions) ([]byte, error) {
	opts, err := opts.normalize()
	if err != nil {
		return nil, err
	}

	payload := pycOrMarshal
	if len(payload) >= len(PycHeader) && bytes.Equal(payload[:len(PycHeader)], PycHeader) {
		payload = payload[len(PycHeader):]
	}

	return MarshalEncrypt(payload, opts)
}

// MarshalEncrypt converts a standard Python marshal object stream into the
// NetEase-encrypted marshal format understood by `RepairMarshal`.
func MarshalEncrypt(marshal []byte, opts EncryptOptions) ([]byte, error) {
	opts, err := opts.normalize()
	if err != nil {
		return nil, err
	}

	s := &encryptState{
		opts:            opts,
		internedStrings: make([][]byte, 0),
		opcodeNameByPy:  invertOpcodeMap(opcodes),
	}

	reader := bytes.NewReader(marshal)
	var out bytes.Buffer
	if err := s.encryptObject(reader, &out); err != nil {
		return nil, err
	}
	if reader.Len() != 0 {
		return nil, fmt.Errorf("trailing bytes after marshal object: %d", reader.Len())
	}
	return out.Bytes(), nil
}

type encryptState struct {
	opts EncryptOptions

	internedStrings [][]byte
	opcodeNameByPy  map[byte]string
}

func invertOpcodeMap(m map[string]byte) map[byte]string {
	out := make(map[byte]string, len(m))
	for name, b := range m {
		out[b] = name
	}
	return out
}

func (s *encryptState) encryptObject(reader *bytes.Reader, writer *bytes.Buffer) error {
	tagType, err := reader.ReadByte()
	if err != nil {
		return err
	}

	switch tagType {
	case TYPE_NULL:
		// Marshal "null object" sentinel used to terminate dicts.
		if err := writer.WriteByte(TYPE_NULL); err != nil {
			return err
		}
		return ErrNull
	case TYPE_NONE, TYPE_STOPITER, TYPE_ELLIPSIS, TYPE_FALSE, TYPE_TRUE:
		return writer.WriteByte(tagType)
	case TYPE_INT:
		if err := writer.WriteByte(TYPE_INT); err != nil {
			return err
		}
		return copyBytes(reader, writer, 4)
	case TYPE_INT64:
		if err := writer.WriteByte(TYPE_INT64); err != nil {
			return err
		}
		return copyBytes(reader, writer, 8)
	case TYPE_LONG:
		if err := writer.WriteByte(TYPE_LONG); err != nil {
			return err
		}
		return copyPyLong(reader, writer)
	case TYPE_FLOAT:
		if err := writer.WriteByte(TYPE_FLOAT); err != nil {
			return err
		}
		return copyFloat(reader, writer)
	case TYPE_BINARY_FLOAT:
		if err := writer.WriteByte(TYPE_BINARY_FLOAT); err != nil {
			return err
		}
		return copyBytes(reader, writer, 8)
	case TYPE_COMPLEX:
		if err := writer.WriteByte(TYPE_COMPLEX); err != nil {
			return err
		}
		if err := copyFloat(reader, writer); err != nil {
			return err
		}
		return copyFloat(reader, writer)
	case TYPE_BINARY_COMPLEX:
		if err := writer.WriteByte(TYPE_BINARY_COMPLEX); err != nil {
			return err
		}
		return copyBytes(reader, writer, 16)

	case TYPE_STRING:
		return s.encryptStringLike(reader, writer, false, false)
	case TYPE_UNICODE:
		return s.encryptStringLike(reader, writer, true, false)
	case TYPE_INTERNED:
		return s.encryptStringLike(reader, writer, false, true)
	case TYPE_STRINGREF:
		return s.encryptStringRef(reader, writer)

	case TYPE_TUPLE, TYPE_LIST, TYPE_SET, TYPE_FROZENSET:
		if err := writer.WriteByte(tagType); err != nil {
			return err
		}
		return s.encryptListLike(reader, writer)
	case TYPE_DICT:
		if err := writer.WriteByte(TYPE_DICT); err != nil {
			return err
		}
		return s.encryptDict(reader, writer)

	case TYPE_CODE:
		return s.encryptCodeObject(reader, writer)

	default:
		return fmt.Errorf("unsupported marshal tag: %x", tagType)
	}
}

func (s *encryptState) encryptListLike(reader *bytes.Reader, writer *bytes.Buffer) error {
	var n int32
	if err := binary.Read(reader, binary.LittleEndian, &n); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, n); err != nil {
		return err
	}

	for i := int32(0); i < n; i++ {
		if err := s.encryptObject(reader, writer); err != nil {
			return err
		}
	}
	return nil
}

func (s *encryptState) encryptDict(reader *bytes.Reader, writer *bytes.Buffer) error {
	for {
		next, err := reader.ReadByte()
		if err != nil {
			return err
		}
		if next == TYPE_NULL {
			if err := writer.WriteByte(TYPE_NULL); err != nil {
				return err
			}
			return nil
		}
		if err := reader.UnreadByte(); err != nil {
			return err
		}

		if err := s.encryptObject(reader, writer); err != nil {
			return err
		}
		if err := s.encryptObject(reader, writer); err != nil {
			if errors.Is(err, ErrNull) {
				return fmt.Errorf("invalid marshal dict: null value")
			}
			return err
		}
	}
}

func (s *encryptState) encryptStringRef(reader *bytes.Reader, writer *bytes.Buffer) error {
	var idx int32
	if err := binary.Read(reader, binary.LittleEndian, &idx); err != nil {
		return err
	}
	if idx < 0 || int(idx) >= len(s.internedStrings) {
		return fmt.Errorf("stringref out of range: %d", idx)
	}
	return s.writeNetEaseString(writer, s.internedStrings[idx], false, false)
}

func (s *encryptState) encryptStringLike(reader *bytes.Reader, writer *bytes.Buffer, isUnicode bool, isInterned bool) error {
	var n int32
	if err := binary.Read(reader, binary.LittleEndian, &n); err != nil {
		return err
	}
	if n < 0 {
		return fmt.Errorf("negative string length: %d", n)
	}
	data := make([]byte, n)
	if _, err := io.ReadFull(reader, data); err != nil {
		return err
	}

	if isInterned {
		s.internedStrings = append(s.internedStrings, data)
	}

	return s.writeNetEaseString(writer, data, isUnicode, isInterned)
}

func (s *encryptState) writeNetEaseString(writer *bytes.Buffer, plain []byte, isUnicode bool, isInterned bool) error {
	tag, err := s.netEaseStringTag(isUnicode, isInterned)
	if err != nil {
		return err
	}
	if err := writer.WriteByte(tag); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, int32(len(plain))); err != nil {
		return err
	}

	encrypted := make([]byte, len(plain))
	copy(encrypted, plain)

	switch s.opts.StringVersion {
	case 0:
		for i := 0; i < len(encrypted); i++ {
			encrypted[i] ^= 0x8d
		}
	case 1:
		c, err := rc4.NewCipher([]byte{0xA7, 0x0D, 0x37, 0x7A})
		if err != nil {
			return err
		}
		c.XORKeyStream(encrypted, encrypted)
	case 2:
		c, err := rc4.NewCipher([]byte{0x8d, 0x06, 0xe8, 0xc8, 0xb7, 0xd7, 0xb7, 0x28, 0x46, 0x51, 0xae, 0x04})
		if err != nil {
			return err
		}
		c.XORKeyStream(encrypted, encrypted)
	default:
		return ErrUnsupportedStringVersion
	}

	_, err = writer.Write(encrypted)
	return err
}

func (s *encryptState) netEaseStringTag(isUnicode bool, isInterned bool) (byte, error) {
	switch s.opts.StringVersion {
	case 0:
		switch {
		case isUnicode:
			return TYPE_NETEASE_UNICODE_0, nil
		case isInterned:
			return TYPE_NETEASE_INTERNED_0, nil
		default:
			return TYPE_NETEASE_STRING_0, nil
		}
	case 1:
		switch {
		case isUnicode:
			return TYPE_NETEASE_UNICODE_1, nil
		case isInterned:
			return TYPE_NETEASE_INTERNED_1, nil
		default:
			return TYPE_NETEASE_STRING_1, nil
		}
	case 2:
		switch {
		case isUnicode:
			return TYPE_NETEASE_UNICODE_2, nil
		case isInterned:
			return TYPE_NETEASE_INTERNED_2, nil
		default:
			return TYPE_NETEASE_STRING_2, nil
		}
	default:
		return 0, ErrUnsupportedStringVersion
	}
}

func (s *encryptState) encryptCodeObject(reader *bytes.Reader, writer *bytes.Buffer) error {
	argcount := make([]byte, 4)
	locals := make([]byte, 4)
	stacksize := make([]byte, 4)
	flags := make([]byte, 4)

	if _, err := io.ReadFull(reader, argcount); err != nil {
		return fmt.Errorf("read code.argcount: %w", err)
	}
	if _, err := io.ReadFull(reader, locals); err != nil {
		return fmt.Errorf("read code.locals: %w", err)
	}
	if _, err := io.ReadFull(reader, stacksize); err != nil {
		return fmt.Errorf("read code.stacksize: %w", err)
	}
	if _, err := io.ReadFull(reader, flags); err != nil {
		return fmt.Errorf("read code.flags: %w", err)
	}

	codePlain, err := s.readPlainMarshalString(reader)
	if err != nil {
		return fmt.Errorf("read code bytes: %w", err)
	}
	codeNetEase, err := s.encodeOpcodeBytes(codePlain, s.opts.OpcodeType)
	if err != nil {
		return err
	}
	var codeBuf bytes.Buffer
	if err := s.writeNetEaseString(&codeBuf, codeNetEase, false, false); err != nil {
		return err
	}

	var consts bytes.Buffer
	if err := s.encryptObject(reader, &consts); err != nil {
		return err
	}
	var names bytes.Buffer
	if err := s.encryptObject(reader, &names); err != nil {
		return err
	}
	var varnames bytes.Buffer
	if err := s.encryptObject(reader, &varnames); err != nil {
		return err
	}
	var freevars bytes.Buffer
	if err := s.encryptObject(reader, &freevars); err != nil {
		return err
	}
	var cellvars bytes.Buffer
	if err := s.encryptObject(reader, &cellvars); err != nil {
		return err
	}
	var filename bytes.Buffer
	if err := s.encryptObject(reader, &filename); err != nil {
		return err
	}
	var name bytes.Buffer
	if err := s.encryptObject(reader, &name); err != nil {
		return err
	}

	firstlineno := make([]byte, 4)
	if _, err := io.ReadFull(reader, firstlineno); err != nil {
		return fmt.Errorf("read code.firstlineno: %w", err)
	}

	var lnotab bytes.Buffer
	if err := s.encryptObject(reader, &lnotab); err != nil {
		return err
	}

	opcodeTypeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(opcodeTypeBytes, s.opts.OpcodeType)

	switch s.opts.CodeVersion {
	case 0:
		if err := writer.WriteByte(TYPE_NETEASE_CODE_0); err != nil {
			return err
		}
		writer.Write(lnotab.Bytes())
		writer.Write(varnames.Bytes())
		writer.Write(flags)
		writer.Write(freevars.Bytes())
		writer.Write(cellvars.Bytes())
		writer.Write(filename.Bytes())
		writer.Write(stacksize)
		writer.Write(firstlineno)
		writer.Write(consts.Bytes())
		writer.Write(argcount)
		writer.Write(codeBuf.Bytes())
		writer.Write(locals)
		writer.Write(name.Bytes())
		writer.Write(names.Bytes())
		_, err := writer.Write(opcodeTypeBytes)
		return err

	case 1:
		if err := writer.WriteByte(TYPE_NETEASE_CODE_1); err != nil {
			return err
		}
		writer.Write(locals)
		writer.Write(flags)
		writer.Write(consts.Bytes())
		writer.Write(stacksize)
		writer.Write(varnames.Bytes())
		writer.Write(argcount)
		writer.Write(freevars.Bytes())
		writer.Write(names.Bytes())
		writer.Write(cellvars.Bytes())
		writer.Write(name.Bytes())
		writer.Write(codeBuf.Bytes())
		writer.Write(firstlineno)
		writer.Write(lnotab.Bytes())
		writer.Write(opcodeTypeBytes)
		_, err := writer.Write(filename.Bytes())
		return err

	case 2:
		if err := writer.WriteByte(TYPE_NETEASE_CODE_2); err != nil {
			return err
		}
		writer.Write(argcount)
		writer.Write(lnotab.Bytes())
		writer.Write(cellvars.Bytes())
		writer.Write(firstlineno)
		writer.Write(varnames.Bytes())
		writer.Write(consts.Bytes())
		writer.Write(name.Bytes())
		writer.Write(stacksize)
		writer.Write(freevars.Bytes())
		writer.Write(names.Bytes())
		writer.Write(codeBuf.Bytes())
		writer.Write(flags)
		writer.Write(filename.Bytes())
		writer.Write(locals)
		_, err := writer.Write(opcodeTypeBytes)
		return err

	default:
		return ErrUnsupportedCodeVersion
	}
}

func (s *encryptState) readPlainMarshalString(reader *bytes.Reader) ([]byte, error) {
	tag, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch tag {
	case TYPE_STRING, TYPE_UNICODE, TYPE_INTERNED:
		var n int32
		if err := binary.Read(reader, binary.LittleEndian, &n); err != nil {
			return nil, err
		}
		if n < 0 {
			return nil, fmt.Errorf("negative string length: %d", n)
		}
		data := make([]byte, n)
		if _, err := io.ReadFull(reader, data); err != nil {
			return nil, err
		}
		if tag == TYPE_INTERNED {
			s.internedStrings = append(s.internedStrings, data)
		}
		return data, nil
	case TYPE_STRINGREF:
		var idx int32
		if err := binary.Read(reader, binary.LittleEndian, &idx); err != nil {
			return nil, err
		}
		if idx < 0 || int(idx) >= len(s.internedStrings) {
			return nil, fmt.Errorf("stringref out of range: %d", idx)
		}
		return s.internedStrings[idx], nil
	default:
		return nil, fmt.Errorf("expected string-like object, got tag %x", tag)
	}
}

func (s *encryptState) encodeOpcodeBytes(pyBytecode []byte, opcodeType uint32) ([]byte, error) {
	opcodeMap, ok := neteaseOpcodes[opcodeType]
	if !ok {
		return nil, fmt.Errorf("no opcode map found: %x", opcodeType)
	}

	nameToNet := make(map[string]byte, len(opcodeMap))
	for b, name := range opcodeMap {
		if name == "" {
			continue
		}
		nameToNet[name] = b
	}

	out := make([]byte, len(pyBytecode))
	copy(out, pyBytecode)

	for i := 0; i < len(out); {
		pyOp := out[i]
		name, ok := s.opcodeNameByPy[pyOp]
		if !ok {
			return nil, fmt.Errorf("unknown python opcode: %d", pyOp)
		}
		netOp, ok := nameToNet[name]
		if !ok {
			return nil, fmt.Errorf("no netease opcode for %q (python=%d)", name, pyOp)
		}
		out[i] = netOp

		if pyOp >= 90 {
			i += 3
		} else {
			i++
		}
	}

	return out, nil
}
