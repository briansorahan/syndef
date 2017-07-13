package defdiff

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/scgolang/sc"
)

// Do diffs two synthdefs.
// If the returned slice is empty it means the synthdefs are structurally identical.
func Do(s1, s2 *sc.Synthdef) ([][2]string, error) {
	dfr := &differ{}
	return dfr.do(s1, s2)
}

// differ creates a diff of two synthdefs.
type differ struct {
	s1 Synthdef
	s2 Synthdef
}

// crawl crawls the ugen graph, starting at the specified ugens in each synthdef,
// looking for structural differences.
func (d *differ) crawl(diffs [][2]string, idx1, idx2 int) [][2]string {
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
		if IsConstant(in1) && IsConstant(in2) {
			if v1, v2 := d.s1.Constants[oi1], d.s2.Constants[oi2]; v1 != v2 {
				diffs = append(diffs, [2]string{
					fmt.Sprintf("%s (ugen %d), input %d has constant value %f", u1.Name, idx1, i, v1),
					fmt.Sprintf("%s (ugen %d), input %d has constant value %f", u2.Name, idx2, i, v2),
				})
			}
			continue
		}
		if IsConstant(in1) && !IsConstant(in2) {
			diffs = append(diffs, [2]string{
				fmt.Sprintf("%s(%d), input %d is constant (%f)", u1.Name, idx1, i, d.s1.Constants[oi1]),
				fmt.Sprintf("%s(%d), input %d points to %s(%d)", u2.Name, idx2, i, d.s2.Ugens[ui2].Name, ui2),
			})
			continue
		}
		if !IsConstant(in1) && IsConstant(in2) {
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

// do performs a diff.
// The diff shows whether one ugen graph differs structurally from another.
// If the returned slice is empty it means the synthdefs are structurally identical.
func (d *differ) do(s1, s2 *sc.Synthdef) ([][2]string, error) {
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
	return d.crawl([][2]string{}, d1.Root(), d2.Root()), nil
}

// getDefs gets unmarshals to our synthdef representation.
func (d *differ) getDefs(s1, s2 *sc.Synthdef) (d1 Synthdef, d2 Synthdef, err error) {
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
	d.s1, d.s2 = d1, d2
	return d1, d2, nil
}
