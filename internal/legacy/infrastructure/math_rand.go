package infrastructure

import (
	"math/rand/v2"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateMathRand() core.IRand {
	return MathRand{}
}

type MathRand struct{}

func (MathRand) GetRand(n int) int {
	return rand.IntN(n)
}
