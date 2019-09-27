package utils

import (
	"testing"
)

func TestSplitWork(t *testing.T) {
	for i := 1; i < 10; i++ {
		totalWorkCount := 17 * i
		visit := make([]bool, totalWorkCount)
		work := func(index int) {
			visit[index] = !visit[index]
		}
		SplitWork(work, i, totalWorkCount)
		for _, value := range visit {
			if !value {
				t.FailNow()
			}
		}
	}

}
