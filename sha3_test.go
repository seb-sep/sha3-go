package sha3go

import (
	"fmt"
	"reflect"
	"testing"
)

/*
func TestSHAStrings(t *testing.T) {
	fmt.Printf("Hash of string \"%s\": %s\n", "Hello world!", SHA256("Hello world"))
}
*/

func TestBytesToBits(t *testing.T) {
	assertEquals(t, bytesToBits([]byte{0b1011, 0b10, 0b1001}), []byte{1, 0, 1, 1, 1, 0, 1, 0, 0, 1})
}

func assertEquals[T any](t *testing.T, val, expect T) {
	if reflect.DeepEqual(val, expect) {
		t.Log("Values equal: ", expect)
	} else {
		t.Error("Values not equal: ", val, " given, ", expect, " expected\n")
	}
}

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
