package mcpk

import (
	"math"
)

type RotorObj struct {
	key       [5]int16
	seed      [3]int32
	size      int
	sizeMask  int
	rotors    int
	eRotor    []byte
	dRotor    []byte
	positions []byte
	advances  []byte

	isInited bool
}

func (r *RotorObj) setKey(key string) {
	k1, k2, k3, k4, k5 := uint64(995), uint64(576), uint64(767), uint64(671), uint64(463)

	for i := 0; i < len(key); i++ {
		ki := uint64(key[i]) & 0xFF

		k1 = ((k1 << 3) | (k1 >> 13) + ki) & 65535
		k2 = ((k2 << 3) | (k2 >> 13) ^ ki) & 65535
		k3 = ((k3 << 3) | (k3 >> 13) - ki) & 65535
		k4 = (ki - ((k4 << 3) | (k4 >> 13))) & 65535
		k5 = ((k5 << 3) | (k5 >> 13) ^ (^ki & 0xFFFF)) & 65535
	}

	r.key[0] = int16(k1)
	r.key[1] = int16(k2 | 1)
	r.key[2] = int16(k3)
	r.key[3] = int16(k4)
	r.key[4] = int16(k5)

	r.setSeed()
}

func (r *RotorObj) setSeed() {
	r.seed[0] = int32(r.key[0])
	r.seed[1] = int32(r.key[1])
	r.seed[2] = int32(r.key[2])
	r.isInited = false
}

// rRandom returns the next random number in the range [0.0, 1.0)
func (r *RotorObj) rRandom() float64 {
	x, y, z := r.seed[0], r.seed[1], r.seed[2]

	x = 171*(x%177) - 2*(x/177)
	y = 172*(y%176) - 35*(y/176)
	z = 170*(z%178) - 63*(z/178)

	if x < 0 {
		x = x + 30269
	}
	if y < 0 {
		y = y + 30307
	}
	if z < 0 {
		z = z + 30323
	}

	r.seed[0] = x
	r.seed[1] = y
	r.seed[2] = z

	term := float64(x)/30269.0 + float64(y)/30307.0 + float64(z)/30323.0
	val := term - math.Floor(term)

	if val >= 1.0 {
		val = 0.0
	}

	return val
}

// rRand returns a random int16 between 0 and s-1, inclusive
func (r *RotorObj) rRand(s int16) int16 {
	return int16(int(math.Floor(r.rRandom()*float64(s))) % int(s))
}

func NewRotorObj(numRotors int, key string) *RotorObj {
	r := &RotorObj{}

	r.setKey(key)

	r.size = 256
	r.sizeMask = r.size - 1
	r.rotors = numRotors

	r.eRotor = make([]byte, numRotors*r.size)
	r.dRotor = make([]byte, numRotors*r.size)
	r.positions = make([]byte, numRotors)
	r.advances = make([]byte, numRotors)

	return r
}

// rtrMakeIDRotor sets ROTOR to the identity permutation.
func (r *RotorObj) MakeIDRotor(rtr []byte) {
	for j := 0; j < r.size; j++ {
		rtr[j] = byte(j)
	}
}

// rtrERotors sets the current set of encryption rotors.
func (r *RotorObj) ERotors() {
	for i := 0; i < r.rotors; i++ {
		r.MakeIDRotor(r.eRotor[i*r.size:])
	}
}

// rtrDRotors sets the current set of decryption rotors.
func (r *RotorObj) DRotors() {
	for i := 0; i < r.rotors; i++ {
		for j := 0; j < r.size; j++ {
			r.dRotor[i*r.size+j] = byte(j)
		}
	}
}

// rtrPositions sets the positions of the rotors at this time.
func (r *RotorObj) Positions() {
	for i := 0; i < r.rotors; i++ {
		r.positions[i] = 1
	}
}

// rtrAdvances sets the number of positions to advance the rotors at a time.
func (r *RotorObj) Advances() {
	for i := 0; i < r.rotors; i++ {
		r.advances[i] = 1
	}
}

// rtrPermuteRotor permutes the E rotor and makes the D rotor its inverse.
func (r *RotorObj) PermuteRotor(e, d []byte) {
	i := r.size
	r.MakeIDRotor(e)
	for i >= 2 {
		q := r.rRand(int16(i))
		i--
		j := e[q]
		e[q] = e[i]
		e[i] = j
		d[j] = byte(i)
	}
	
	d[e[0]] = 0
}

// rtrInit initializes the rotor machine given a key.
func (r *RotorObj) Init() {
	r.setSeed()
	r.Positions()
	r.Advances()
	r.ERotors()
	r.DRotors()
	for i := 0; i < r.rotors; i++ {
		r.positions[i] = byte(r.rRand(int16(r.size)))
		r.advances[i] = byte(1 + 2*(r.rRand(int16(r.size/2))))
		r.PermuteRotor(r.eRotor[i*r.size:], r.dRotor[i*r.size:])
	}
	r.isInited = true
}

// rtrAdvance changes the RTR-positions vector, using the RTR-advances vector.
func (r *RotorObj) advance() {
	var i, temp int
	if r.sizeMask != 0 {
		for i = 0; i < r.rotors; i++ {
			temp = int(r.positions[i]) + int(r.advances[i])
			r.positions[i] = byte(temp & r.sizeMask)
			if temp >= r.size && i < r.rotors-1 {
				r.positions[i+1] = 1 + r.positions[i+1]
			}
		}
	} else {
		for i = 0; i < r.rotors; i++ {
			temp = int(r.positions[i]) + int(r.advances[i])
			r.positions[i] = byte(temp & r.size)
			if temp >= r.size && i < r.rotors-1 {
				r.positions[i+1] = 1 + r.positions[i+1]
			}
		}
	}
}

// rtrEChar encrypts a byte p with the current rotor machine.
func (r *RotorObj) EChar(p byte) byte {
	i := 0
	tp := p
	if r.sizeMask != 0 {
		for i < r.rotors {
			tp = r.eRotor[(i*r.size)+(int(r.positions[i]^tp)&r.sizeMask)]
			i++
		}
	} else {
		for i < r.rotors {
			tp = r.eRotor[(i*r.size)+(int(r.positions[i]^tp)%int(r.size))]
			i++
		}
	}
	r.advance()
	return tp
}

// rtrDChar decrypts a byte c with the current rotor machine.
func (r *RotorObj) DChar(c byte) byte {
	i := r.rotors - 1
	tc := c

	if r.sizeMask != 0 {
		for i >= 0 {
			tc = (r.positions[i] ^ r.dRotor[(i*r.size)+int(tc)]) & byte(r.sizeMask)
			i--
		}
	} else {
		for i >= 0 {
			tc = (r.positions[i] ^ r.dRotor[(i*r.size)+int(tc)]) % byte(r.size)
			i--
		}
	}
	r.advance()
	return tc
}

// rtrERegion performs a rotor encryption of a byte slice data using the given rotor machine.
func (r *RotorObj) ERegion(data []byte, doInit bool) {
	if doInit || !r.isInited {
		r.Init()
	}
	for i := range data {
		data[i] = r.EChar(data[i])
	}
}

// rtrDRegion performs a rotor decryption of a byte slice data using the given rotor machine.
func (r *RotorObj) DRegion(data []byte, doInit bool) {
	if doInit || !r.isInited {
		r.Init()
	}
	for i := range data {
		data[i] = r.DChar(data[i])
	}
}
