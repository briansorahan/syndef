package main

type synthdef struct {
	Constants []float64 `json:"constants"`
	Ugens     []ugen    `json:"ugens"`
}

func (s synthdef) root() int {
	parents := make([]int, len(s.Ugens)) // Number of parents per ugen.
	for _, u := range s.Ugens {
		for _, in := range u.Inputs {
			if isConstant(in) {
				continue
			}
			parents[in.UgenIndex]++
		}
	}
	for i, count := range parents {
		if count == 0 {
			return i
		}
	}
	return 0
}

type ugen struct {
	Inputs       []input `json:"inputs"`
	Name         string  `json:"name"`
	Outputs      []int   `json:"outputs"`
	Rate         int     `json:"rate"`
	SpecialIndex int     `json:"rate"`
}

type input struct {
	OutputIndex int `json:"outputIndex"`
	UgenIndex   int `json:"ugenIndex"` // UgenIndex will be -1 when the input is a constant
}

func isConstant(in input) bool {
	return in.UgenIndex == -1
}

var commutative = map[string]struct{}{
	"BinaryOpUgen": struct{}{},
}
