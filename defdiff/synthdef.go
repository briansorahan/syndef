package defdiff

// Synthdef presents a simplified version of a SuperCollider synthdef.
type Synthdef struct {
	Constants []float64 `json:"constants"`
	Ugens     []Ugen    `json:"ugens"`
}

// Root returns the root node in the synthdef's ugen graph.
func (s Synthdef) Root() int {
	parents := make([]int, len(s.Ugens)) // Number of parents per ugen.
	for _, u := range s.Ugens {
		for _, in := range u.Inputs {
			if IsConstant(in) {
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

type Ugen struct {
	Inputs       []Input `json:"inputs"`
	Name         string  `json:"name"`
	Outputs      []int   `json:"outputs"`
	Rate         int     `json:"rate"`
	SpecialIndex int     `json:"rate"`
}

type Input struct {
	OutputIndex int `json:"outputIndex"`
	UgenIndex   int `json:"ugenIndex"` // UgenIndex will be -1 when the input is a constant
}

func IsConstant(in Input) bool {
	return in.UgenIndex == -1
}

var commutative = map[string]struct{}{
	"BinaryOpUgen": struct{}{},
}
