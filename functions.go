package sha3

import (
	"math"
)

type stateArray [5][5][]byte

var indexMap = [5]int{2, 3, 4, 0, 1}

/*
Convert a binary string to a state array, as specified by Keccak.
*/
func toStateArray(str []byte) stateArray {
	b := len(str)
	w := b / 25
	//l := math.Log2(float64(w))

	state := newStateArray(w)

	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < w; z++ {
				index := w*(5*y+x) + z
				put(state, x, y, z, str[index])
			}
		}
	}
	return state
}

/*
Convert a state array back to a binary string, as specified by Keccak.
*/
func toBinaryString(state stateArray) []byte {
	str := []byte{}
	for j := 0; j < 5; j++ {
		str = append(str, plane(state, j)...)
	}
	return str
}

/*
Get a lane from the state array.
A lane is a concatenation of all bits in a state array of a given row and column.
*/
func lane(state stateArray, i int, j int) []byte {
	return state[indexMap[i]][indexMap[j]]
}

/*
Get a plane from the state array.
A plane is a concatenation of all lanes in the same row.
*/
func plane(state stateArray, j int) []byte {
	str := []byte{}
	for i := 0; i < 5; i++ {
		str = append(str, lane(state, i, j)...)
	}
	return str
}

/*
Create a new state array of the given length w.
*/
func newStateArray(w int) stateArray {
	var state stateArray

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			state[i][j] = make([]byte, w)
		}
	}
	return state
}

/*
Implements the theta algorithm as specified by Keccak. Note that ^ is the XOR operator.
*/
func theta(state stateArray) stateArray {
	w := len(state[0][0])

	//Calculate the C mapping
	var C [5][]byte
	for x := 0; x < 5; x++ {
		for z := 0; z < w; z++ {
			C[x][z] = index(state, x, 0, z) ^
				index(state, x, 1, z) ^
				index(state, x, 2, z) ^
				index(state, x, 3, z) ^
				index(state, x, 4, z)
		}
	}

	//Calculate the D mapping
	var D [5][]byte
	for x := 0; x < 5; x++ {
		for z := 0; z < w; z++ {
			D[x][z] = C[(x-1)%5][z] ^ C[(x+1)%5][(z-1)%w]
		}
	}
	stateTheta := stateArrayMap(w, func(x int, y int, z int) byte { return index(state, x, y, z) ^ D[x][z] })

	/*for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < w; z++ {
				value := index(state, x, y, z) ^ D[x][z]
				put(stateTheta, x, y, z, value)
			}
		}
	}*/

	return stateTheta
}

/*
Implements the rho algorithm as specified by Keccak.
*/
func rho(state stateArray) stateArray {

	w := len(state[0][0])
	stateRho := newStateArray(w)
	for z := 0; z < w; z++ {
		put(stateRho, 0, 0, z, index(state, 0, 0, z))
	}

	x, y := 1, 0
	for t := 0; t <= 23; t++ {
		for z := 0; z < w; z++ {
			put(stateRho, x, y, z, index(state, x, y, (z-(t+1)*(t+2)/2)%w))
		}
		x, y = y, (2*x+3*y)%5
	}

	return stateRho
}

/*
Implements the pi function as specified by Keccak.
*/
func pi(state stateArray) stateArray {
	w := len(state[0][0])
	return stateArrayMap(w, func(x int, y int, z int) byte { return index(state, (x+3*y)%5, x, z) })
}

/*
Implements the chi function as specified by Keccak.
*/
func chi(state stateArray) stateArray {
	w := len(state[0][0])
	return stateArrayMap(w, func(x int, y int, z int) byte {
		return index(state, x, y, z) ^
			(index(state, (x+1)%5, y, z)^1)*index(state, (x+2)%5, y, z)
	})
}

/*
Implements the iota function as specified by Keccak.
*/
func iota(state stateArray, i int) stateArray {
	w := len(state[0][0])
	l := int(math.Log2(float64(w)))
	stateIota := stateArrayMap(w, func(x int, y int, z int) byte { return index(state, x, y, z) })
	//Are Go slices initialized to zero values?
	RC := make([]byte, w)
	for j := 0; j <= l; j++ {
		RC[int(math.Pow(2.0, float64(j)))-1] = rc(j + 7*i)
	}

	for z := 0; z < w; z++ {
		value := index(stateIota, 0, 0, z) ^ RC[z]
		put(stateIota, 0, 0, z, value)
	}
	return stateIota
}

/*
Implements the rc function as specified by Keccak.
*/
func rc(t int) byte {
	if mod := t % 255; mod == 0 {
		return 1
	}
	R := []byte{1, 0, 0, 0, 0, 0, 0, 0}
	for i := 1; i <= 255; i++ {
		R = append([]byte{0}, R...)
		R[0] = R[0] ^ R[8]
		R[4] = R[4] ^ R[8]
		R[5] = R[5] ^ R[8]
		R[6] = R[6] ^ R[8]
		R = R[:7]
	}
	return R[0]
}

/*
Implements the Rnd function as specified by Keccak.
Applies the theta, rho, pi, chi, and iota functions in sequence to a state array.
*/
func rnd(state stateArray, rounds int) stateArray {
	return iota(chi(pi(rho(theta(state)))), rounds)
}

/*
Implements the Keccak-p algorithm.
*/
func KeccakP(str []byte, rounds int) []byte {
	state := toStateArray(str)
	b := len(str)
	w := b / 25
	l := int(math.Log2(float64(w)))
	for i := (12 + 2*l - rounds); i <= (12 + 2*l - 1); i++ {
		state = rnd(state, i)
	}

	//TODO: Assert that input and output strings are the same length?
	return toBinaryString(state)
}

/*
Implements the Keccak-f algorithm.
*/
func KeccakF(str []byte) []byte {
	b := len(str)
	w := b / 25
	l := int(math.Log2(float64(w)))
	return KeccakP(str, 12+2*l)
}

/*
Implements the sponge function as specified by Keccak.
Spongs maps the binary string of arbitrary length to a binary string
of length d.
f is the mapping function of b-length binary strings to b-length binary strings,
pad is the given padding function,
r is the padding rate,
str is the input binary string,
and d is the target length.
*/
func Sponge(f func(str []byte) []byte, b int, pad func(x int, m uint) []byte, r int, str []byte, d uint) []byte {
	bitwiseXOR := func(str1 []byte, str2 []byte) []byte {
		l := len(str1)
		/*if l != len(str2) {
			return []byte{}, fmt.Errorf("Binary strings are not the same length")
		}*/
		result := make([]byte, l)
		for i := 0; i < l; i++ {
			result[i] = str1[i] ^ str2[i]
		}
		return result
	}

	P := append(str, pad(r, uint(len(str)))...)
	n := len(P) / r
	c := b - r
	S := make([]byte, b)
	for i := 0; i <= n-1; i++ {
		S = f(bitwiseXOR(S, append(P[i:i+r], make([]byte, c)...)))
	}
	Z := []byte{}
	for int(d) > len(Z) {
		Z = append(Z, S[0:r]...)
		S = f(S)
	}
	return Z
}

/*
Implements the pad10*1 padding function as specified by Keccak.
As specified, pad(x, m) returns a string such that m + len(pad(x, m))
is a multiple of x.
*/
func pad(x int, m uint) []byte {
	j := (-int(m) - 2) % x
	str := make([]byte, j+2)
	str[0] = 1
	str[j+1] = 1
	return str
}

/*
Populate a state array with the given length with the given mapping function.
*/
func stateArrayMap(w int, l func(int, int, int) byte) stateArray {
	stateNew := newStateArray(w)

	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < w; z++ {
				put(stateNew, x, y, z, l(x, y, z))
			}
		}
	}

	return stateNew
}

/*
Access a value in the state array by the conventions of the Keccak state array.
*/
func index(state stateArray, i int, j int, k int) byte {
	return state[indexMap[i]][indexMap[j]][k]
}

/*
Put a value in the state array by the conventions of the Keccak state array.
*/
func put(state stateArray, i int, j int, k int, value byte) {
	state[indexMap[i]][indexMap[j]][k] = value
}
