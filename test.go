package sha3

import (
	"fmt"
	"testing"
)

func TestStateArray(t *testing.T) {
	printStateArray(toStateArray([]byte{0, 1, 1, 0, 1}))
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
