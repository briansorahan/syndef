package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/scgolang/sc"
)

// differ creates a diff of two synthdefs.
type differ struct {
	s1 synthdef
	s2 synthdef
}

// crawl crawls the ugen graph, starting at the specified ugens in each synthdef,
// looking for structural differences.
func (d differ) crawl(diffs [][2]string, idx1, idx2 int) [][2]string {
	var (
		u1 = d.s1.Ugens[idx1]
		u2 = d.s2.Ugens[idx2]
	)
	if u1.Name != u2.Name {
		return append(diffs, [2]string{
			fmt.Sprintf("ugen %d is a %s", idx1, u1.Name),
			fmt.Sprintf("ugen %d is a %s", idx2, u1.Name),
		})
	}
	if l1, l2 := len(u1.Inputs), len(u2.Inputs); l1 != l2 {
		return append(diffs, [2]string{
			fmt.Sprintf("%s has %d inputs", u1.Name, l1),
			fmt.Sprintf("%s has %d inputs", u2.Name, l2),
		})
	}
	// They have the same number of inputs.
	for i, in1 := range u1.Inputs {
		var (
			in2 = u2.Inputs[i]
			oi1 = in1.OutputIndex
			oi2 = in2.OutputIndex
			ui1 = in1.UgenIndex
			ui2 = in2.UgenIndex
		)
		if isConstant(ui1) && isConstant(ui2) {
			if v1, v2 := d.s1.Constants[oi1], d.s2.Constants[oi2]; v1 != v2 {
				diffs = append(diffs, [2]string{
					fmt.Sprintf("%s (ugen %d), input %d has constant value %f", u1.Name, idx1, i, v1),
					fmt.Sprintf("%s (ugen %d), input %d has constant value %f", u2.Name, idx2, i, v2),
				})
			}
			continue
		}
		if isConstant(ui1) && !isConstant(ui2) {
			diffs = append(diffs, [2]string{
				fmt.Sprintf("%s(%d), input %d is constant (%f)", u1.Name, idx1, i, d.s1.Constants[oi1]),
				fmt.Sprintf("%s(%d), input %d points to %s(%d)", u2.Name, idx2, i, d.s2.Ugens[ui2].Name, ui2),
			})
			continue
		}
		if !isConstant(ui1) && isConstant(ui2) {
			diffs = append(diffs, [2]string{
				fmt.Sprintf("%s(%d), input %d points to %s(%d)", u1.Name, idx1, i, d.s1.Ugens[ui1].Name, ui1),
				fmt.Sprintf("%s(%d), input %d is constant (%f)", u2.Name, idx2, i, d.s2.Constants[oi2]),
			})
			continue
		}
		// They are both not constant.
		// TODO: detect cycles
		diffs = append(diffs, d.crawl(diffs, ui1, ui2)...)
	}
	return diffs
}

// do does the diff, printing details to the provided writer.
// The diff shows whether one ugen graph differs structurally from another.
func (d differ) do(s1, s2 *sc.Synthdef) ([][2]string, error) {
	d1, d2, err := d.getDefs(s1, s2)
	if err != nil {
		return nil, err
	}
	// Early out if they have different numbers of ugens or constants.
	if l1, l2 := len(d1.Ugens), len(d2.Ugens); l1 != l2 {
		return [][2]string{
			{
				fmt.Sprintf("%d ugens", l1),
				fmt.Sprintf("%d ugens", l2),
			},
		}, nil
	}
	if l1, l2 := len(d1.Constants), len(d2.Constants); l1 != l2 {
		return [][2]string{
			{
				fmt.Sprintf("%d constants", l1),
				fmt.Sprintf("%d constants", l2),
			},
		}, nil
	}
	return d.crawl([][2]string{}, d1.root(), d2.root()), nil
}

// getDefs gets unmarshals to our synthdef representation.
func (d differ) getDefs(s1, s2 *sc.Synthdef) (d1 synthdef, d2 synthdef, err error) {
	var (
		buf1 = &bytes.Buffer{}
		buf2 = &bytes.Buffer{}
	)
	if err := s1.WriteJSON(buf1); err != nil {
		return d1, d2, err
	}
	if err := s2.WriteJSON(buf2); err != nil {
		return d1, d2, err
	}
	if err := json.NewDecoder(buf1).Decode(&d1); err != nil {
		return d1, d2, err
	}
	if err := json.NewDecoder(buf2).Decode(&d2); err != nil {
		return d1, d2, err
	}
	return d1, d2, nil
}

type synthdef struct {
	Constants []float64 `json:"constants"`
	Ugens     []ugen    `json:"ugens"`
}

func (s synthdef) root() int {
	parents := make([]int, len(s.Ugens)) // Number of parents per ugen.
	for _, u := range s.Ugens {
		for _, in := range u.Inputs {
			if isConstant(in.UgenIndex) {
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

func isConstant(ugenIndex int) bool {
	return ugenIndex == -1
}

var commutative = map[string]struct{}{
	"BinaryOpUgen": struct{}{},
}
