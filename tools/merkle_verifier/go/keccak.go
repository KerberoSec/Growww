package merkle

import (
	"encoding/binary"
	"math/bits"
)

// rc contains round constants for Keccak-f[1600]
var rc = [24]uint64{
	0x0000000000000001, 0x0000000000008082, 0x800000000000808A, 0x8000000080008000,
	0x000000000000808B, 0x0000000080000001, 0x8000000080008081, 0x8000000000008009,
	0x000000000000008A, 0x0000000000000088, 0x0000000080008009, 0x000000008000000A,
	0x000000008000808B, 0x800000000000008B, 0x8000000000008089, 0x8000000000008003,
	0x8000000000008002, 0x8000000000000080, 0x000000000000800A, 0x800000008000000A,
	0x8000000080008081, 0x8000000000008080, 0x0000000080000001, 0x8000000080008008,
}

// rho contains rotational offsets
var rho = [5][5]uint{
	{0, 36, 3, 41, 18},
	{1, 44, 10, 45, 2},
	{62, 6, 43, 15, 61},
	{28, 55, 25, 21, 56},
	{27, 20, 39, 8, 14},
}

func keccakF1600(state *[25]uint64) {
	var lanes [5][5]uint64
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			lanes[x][y] = state[x+5*y]
		}
	}

	for r := 0; r < 24; r++ {
		// Theta
		var c [5]uint64
		for x := 0; x < 5; x++ {
			c[x] = lanes[x][0] ^ lanes[x][1] ^ lanes[x][2] ^ lanes[x][3] ^ lanes[x][4]
		}
		var d [5]uint64
		for x := 0; x < 5; x++ {
			d[x] = c[(x+4)%5] ^ bits.RotateLeft64(c[(x+1)%5], 1)
		}
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				lanes[x][y] ^= d[x]
			}
		}

		// Rho and Pi
		var b [5][5]uint64
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				b[y][(2*x+3*y)%5] = bits.RotateLeft64(lanes[x][y], int(rho[x][y]))
			}
		}

		// Chi
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				lanes[x][y] = b[x][y] ^ ((^b[(x+1)%5][y]) & b[(x+2)%5][y])
			}
		}

		// Iota
		lanes[0][0] ^= rc[r]
	}

	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			state[x+5*y] = lanes[x][y]
		}
	}
}

// Keccak256 computes standard Ethereum Keccak-256 hash (pure Go, 0 dependencies)
func Keccak256(data []byte) [32]byte {
	rate := 136 // 1088 bits
	var state [25]uint64

	padLen := rate - (len(data) % rate)
	padded := make([]byte, len(data)+padLen)
	copy(padded, data)
	if padLen == 1 {
		padded[len(data)] = 0x81
	} else {
		padded[len(data)] = 0x01
		padded[len(padded)-1] = 0x80
	}

	for offset := 0; offset < len(padded); offset += rate {
		block := padded[offset : offset+rate]
		for i := 0; i < rate/8; i++ {
			lane := binary.LittleEndian.Uint64(block[i*8 : (i+1)*8])
			state[i] ^= lane
		}
		keccakF1600(&state)
	}

	var out [32]byte
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint64(out[i*8:(i+1)*8], state[i])
	}
	return out
}
