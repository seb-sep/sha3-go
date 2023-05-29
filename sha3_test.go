package sha3go

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSHA(t *testing.T) {
	assertEquals(t, sha3([]byte{1, 35, 235, 50, 103}, 256), []byte{1, 35, 235, 50, 103})
}

func TestBytesToBits(t *testing.T) {
	assertEquals(t, bytesToBits([]byte{0b1011, 0b10, 0b1001}), []byte{0, 0, 0, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 1})
}

func TestBitsToBytes(t *testing.T) {
	assertEquals(t, bitsToBytes([]byte{0, 1, 0, 1, 0, 1, 0, 0, 1}), []byte{0b0, 0b10101001})
	assertEquals(t, bitsToBytes([]byte{0, 0, 0, 1, 0, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0}), []byte{0b10101, 0b00110010})
	assertEquals(t, bitsToBytes([]byte{0, 1, 0, 1, 0, 1, 0, 0, 1, 0, 0, 1, 1, 0, 1, 0, 1, 0, 0, 1, 1, 1, 0, 1, 0, 0, 1, 0, 0, 1, 1, 1}), []byte{0b1010100, 0b10011010, 0b10011101, 0b100111})
}

func assertEquals[T any](t *testing.T, val, expect T) {
	if reflect.DeepEqual(val, expect) {
		t.Log("Values equal: ", expect)
	} else {
		t.Error("Values not equal: ", val, " given, ", expect, " expected\n")
	}
}

func TestBitsToByte(t *testing.T) {
	assertEquals(t, bitsToByte([]byte{1, 0, 1, 0}), 0b01010)
	assertEquals(t, bitsToByte([]byte{1, 1, 0, 1, 1, 0, 0, 1}), 0b11011001)
}
func bitsToByte(bits []byte) byte {
	len := len(bits)
	assert(len <= 8, "bitsToByte: bit string must be 8 bits or less")
	var ans byte
	for i := len - 1; i >= 0; i-- {
		ans += bits[len-1-i] << i
	}
	return ans
}

/*
func TestByteToBits(t *testing.T) {
	assertEquals(t, byteToBits(0b1011011), []byte{0, 1, 0, 1, 1, 0, 1, 1})
	assertEquals(t, byteToBits(0b0), []byte{0, 0, 0, 0, 0, 0, 0, 0})
}
func byteToBits(b byte) []byte {
	bits := make([]byte, 8)
	for i := 0; i < 8; i++ {
		bits[i] = (b >> (7 - i)) & 0b00000001
	}
	return bits
}
*/
/*
Print out a state array by slices.
*/
func printStateArray(state stateArray) {
	for k, _ := range state[0][0] {
		for i, _ := range state {
			for j, _ := range state[i] {
				fmt.Printf("%b ", state[i][j][k])
			}
		}
	}
}
