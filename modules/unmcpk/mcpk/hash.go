package mcpk

import "encoding/binary"

// replicates _Z8StringIDPKc
// https://github.com/elnx/0x9BE74448
func StringID(input string) uint32 {
	var eax, ebx, ecx, edx, edi, esi uint32
	var num uint64

	var v uint32
	var i int32
	m := make([]uint32, 0x46)
	buffer := make([]byte, 0x100)

	for i = 0; i < int32(len(input)); i++ {
		buffer[i] = input[i]
	}

	length := len(input) / 4
	if len(input)%4 != 0 {
		length++
	}
	for i = 0; i < int32(length); i++ {
		m[i] = binary.LittleEndian.Uint32(buffer[i*4:])
	}

	m[i] = 0x9BE74448
	i++
	m[i] = 0x66F42C48
	i++

	v = 0xF4FA8928

	edi = 0x7758B42B
	esi = 0x37A8470E

	for ecx = 0; ecx < uint32(i); ecx++ {
		ebx = 0x267B0B11
		v = (v << 1) | (v >> 0x1F)
		ebx ^= v
		eax = m[ecx]
		esi ^= eax
		edi ^= eax
		edx = ebx
		edx += edi
		edx |= 0x02040801
		edx &= 0xBFEF7FDF
		num = uint64(edx)
		num *= uint64(esi)
		eax = uint32(num)
		edx = uint32(num >> 0x20)
		if edx != 0 {
			eax++
		}
		num = uint64(eax)
		num += uint64(edx)
		eax = uint32(num)
		if uint32(num>>0x20) != 0 {
			eax++
		}
		edx = ebx
		edx += esi
		edx |= 0x00804021
		edx &= 0x7DFEFBFF
		esi = eax
		num = uint64(edi)
		num *= uint64(edx)
		eax = uint32(num)
		edx = uint32(num >> 0x20)
		num = uint64(edx)
		num += uint64(edx)
		edx = uint32(num)
		if uint32(num>>0x20) != 0 {
			eax++
		}
		num = uint64(eax)
		num += uint64(edx)
		eax = uint32(num)
		if uint32(num>>0x20) != 0 {
			eax += 2
		}
		edi = eax
	}
	esi ^= edi
	v = esi

	return v
}
