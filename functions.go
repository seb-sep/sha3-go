package sha3go

import (
	"math"
)

// A state array is a 5x5xw array of bits, where w is 1/25 the length of the input binary string.
type stateArray [5][5][]byte

// Used as a mapping function from the sequential order of array indices to the order
// specified by Keccak for the x- and y-axes of the state array.
var indexMap = [5]int{2, 3, 4, 0, 1}

/*
Convert a binary string to a state array, as specified by Keccak.
Essentially, the binary input string is broken into 25 equal pieces which are then rearranged into a 3x3
array of 5 rows and 5 columns.
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
Compress a binary string into a true byte string for type conversions.
We know that this binary string must be of at least length 25, so
an edge case of a string size < 8 need not be handled.
*/
func bitsToBytes(bits []byte) []byte {
	//input slice must be of length 8 or less
	bitsToByte := func(bits []byte) byte {
		len := len(bits)
		assert(len <= 8, "bitsToByte: bit string must be 8 bits or less")
		var ans byte
		for i := len - 1; i >= 0; i-- {
			ans += bits[len-1-i] << i
		}
		return ans
	}

	bitLen := len(bits)
	extraBits := bitLen % 8
	byteLen := bitLen / 8
	bytes := make([]byte, byteLen)
	j := 0
	if extraBits != 0 {
		byteLen++
		j++
		bytes[0] = bitsToByte(bits[:extraBits])
		bytes = append(bytes, 0)
	}

	for i := extraBits; i <= bitLen-8; i += 8 {
		bytes[j] = bitsToByte(bits[i : i+8])
	}

	return bytes
}

/*
Expand a byte string into a binary string for use w/Keccak.
*/
func bytesToBits(bytes []byte) []byte {
	byteToBits := func(b byte) []byte {
		bits := make([]byte, 8)
		for i := 0; i < 8; i++ {
			bits[i] = (b>>i - 7) & 0b00000001
		}
		return bits
	}
	byteLen := len(bytes)
	bitLen := byteLen * 8

	bits := make([]byte, bitLen)

	for i := 0; i < byteLen; i++ {
		bits = append(bits, byteToBits(bytes[i])...)
	}
	return bits
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
For each cell in the state array, theta selects two columns: one with the same z-value and a smaller x-value by 1,
and a column with one less z-value and one greater x-value (all values wrapping around). Theta calculates the parities
of the columns by XORing the elements of the columns, XORing the parities together, and then XORing that result with
the cell's value.
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
Implements the rho algorithm as specified by Keccak. Rho maintains the x- and y-values of each cell in the state
array, but increments the z-value by varying offsets which wrap around w.
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
Where rho preserves the x- and y-values of cells, pi preserves the z-values of cells and rotates the x- and y-values.
Picture a sort of rotation of each slice in the state array about its center (but not a true 90-degree rotation).
*/
func pi(state stateArray) stateArray {
	w := len(state[0][0])
	return stateArrayMap(w, func(x int, y int, z int) byte { return index(state, (x+3*y)%5, x, z) })
}

/*
Implements the chi function as specified by Keccak. Chi takes each cell in the state array, and XORs it with
two other cells in its row: specifically, those with x-values greater by one and two, wrapping around.
*/
func chi(state stateArray) stateArray {
	w := len(state[0][0])
	return stateArrayMap(w, func(x int, y int, z int) byte {
		return index(state, x, y, z) ^
			(index(state, (x+1)%5, y, z)^1)*index(state, (x+2)%5, y, z)
	})
}

/*
Implements the iota function as specified by Keccak. Iota XORs each bit of the center lane (Lane(0,0)) by an output
of rc() which depends on the round index i.
Note: underscore is to prevent collisions with Go's iota keyword.
*/
func iota_(state stateArray, i int) stateArray {
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
Implements the rc function as specified by Keccak. rc takes an integer and calculates a bit based off
that integer. For a certain number of iterations, rc appends a 0 to the front of an initial binary array, XORs various
values in the array with the ninth value in the array, and then truncates that last value. After all iterations have
completed, rc returns the first bit in the array.
*/
func rc(t int) byte {
	if t%255 == 0 {
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
	return iota_(chi(pi(rho(theta(state)))), rounds)
}

/*
Implements the Keccak-p algorithm. Keccak-p converts the string to a state array,
applies the Rnd function to it for the specified number of iterations, and then converts the state array
back into a binary string.
*/
func KeccakP(b int, rounds int) func(str []byte) []byte {

	return func(str []byte) []byte {
		state := toStateArray(str)
		assert(len(str) == b, "Length of string S is not b")
		w := b / 25
		l := int(math.Log2(float64(w)))
		for i := (12 + 2*l - rounds); i <= (12 + 2*l - 1); i++ {
			state = rnd(state, i)
		}

		//TODO: Assert that input and output strings are the same length?
		return toBinaryString(state)
	}
}

/*
Keccak-f is a special case of Keccak-p in which the number of rounds = 12 + 2l, where l is the binary log of
the state array width.
*/
func KeccakF(b int) func(str []byte) []byte {
	w := b / 25
	l := int(math.Log2(float64(w)))
	return KeccakP(b, 12+2*l)
}

/*
Implements the sponge construction as specified by Keccak.
Sponge returns a function mapping the binary string of arbitrary length to a binary string
of length d.
f is the mapping function of b-length binary strings to b-length binary strings,
pad is the given padding function,
r is the padding rate,
str is the input binary string,
and d is the target length.

Sponge first concatenates the appropriate padding onto the input string, splits it into substrings of length r,
and initializes a zero string S of length b. Then, for all the substrings, it concatenates each substring with a
zero string of length c, XORs it with S,runs f on that, and stores the result of f in S. Sponge then initializes
and empty string Z, and until the length of Z is greater than or equal to d, Sponge concatenates the first r bits of
S to Z and reruns f on S. Finally, Sponge returns the first d bits of Z.

Note: While b is not passed in to paramaterize the sponge construction in the Keccak spec, it is necessary
here since b cannot be programatically inferred from the choice of f.
*/
func Sponge(f func(str []byte) []byte, b int, pad func(x int, m uint) []byte, r int) func(str []byte, d uint) []byte {
	bitwiseXOR := func(str1 []byte, str2 []byte) []byte {
		l := len(str1)
		assert(l != len(str2), "Binary strings are not the same length")
		result := make([]byte, l)
		for i := 0; i < l; i++ {
			result[i] = str1[i] ^ str2[i]
		}
		return result
	}

	return func(str []byte, d uint) []byte {
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

}

/*
Keccak denotes the family of sponge functions which use a permutation of Keccak-p[b, 12+2l]
and pad10*1. Keccak[c] represents a special case of Keccak in which b=1600.
*/
func KeccakC(c int) func(str []byte, d uint) []byte {
	return Sponge(KeccakP(1600, 24), 1600, pad, 1600-c)
}

/*
The SHA-3 cryptographic hash function, where len is the desired length of the output.
Restricted to 224, 256, 384, and 512-bit output lengths.
*/
func sha3(str []byte, len uint) []byte {
	return KeccakC(int(len)*2)(append(str, 0, 1), len)
}

func SHA256(str string) string {
	binString := bytesToBits([]byte(str))
	output := sha3(binString, 256)
	return string(bitsToBytes(output))
}

/*
SHA-3 extendable output functions, where the given output length is arbitrary.
n must be one of 128 or 256, for SHAKE128 and SHAKE256 respectively.
*/
func SHAKE(n int, str []byte, len uint) []byte {
	return KeccakC(n*2)(append(str, 1, 1, 1, 1), len)
}

/*
SHA-3 extendable output functions, where the given output length is arbitrary.
n must be one of 128 or 256, for SHAKE128 and SHAKE256 respectively.
*/
func RawSHAKE(n int, str []byte, len uint) []byte {
	return KeccakC(n*2)(append(str, 1, 1), len)
}

/*
Implements the pad10*1 padding function as specified by Keccak.
As specified, pad(x, m) returns a string such that m + len(pad(x, m))
is a multiple of x. The output string is in the shape 10*1, hence the name.
*/
func pad(x int, m uint) []byte {
	j := (-int(m) - 2) % x
	str := make([]byte, j+2)
	str[0] = 1
	str[j+1] = 1
	assert((len(str)+int(m))%x == 0, "padding is improper length")
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

// Runtime boolean assertion similar to assert() in C.
func assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}
