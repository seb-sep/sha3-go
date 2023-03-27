package sha3

import (
	"math"
)

type stateArray [5][5][]byte

func toStateArray(str []byte) stateArray {
	b := len(str)
	w := b / 25
	l := math.Log2(float64(w))

	var state stateArray

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			state[i][j] = make([]byte, w)
		}
	}

}
