package compliance

import "sort"

// Change is one control whose status differs between two assessments.
type Change struct {
	ControlID string
	Title     string
	Severity  Severity
	Before    Status // empty if the control wasn't in the earlier assessment
	After     Status // empty if the control isn't in the later one
}

// Direction classifies a change for display: a pass that became a fail is
// a regression; the reverse an improvement; anything involving unknown or
// not-applicable, or a control added/removed between engine versions, is
// neither — it's a change in what could be measured, not in posture.
func (c Change) Direction() string {
	switch {
	case c.Before == StatusPass && c.After == StatusFail:
		return "regressed"
	case c.Before == StatusFail && c.After == StatusPass:
		return "improved"
	default:
		return "changed"
	}
}

// Diff compares two assessments' findings by control ID and returns the
// controls whose status changed, regressions first, then by severity.
func Diff(before, after Assessment) []Change {
	b := make(map[string]Finding, len(before.Findings))
	for _, f := range before.Findings {
		b[f.ControlID] = f
	}
	a := make(map[string]Finding, len(after.Findings))
	for _, f := range after.Findings {
		a[f.ControlID] = f
	}

	var out []Change
	for id, af := range a {
		bf, ok := b[id]
		if ok && bf.Status == af.Status {
			continue
		}
		c := Change{ControlID: id, Title: af.Title, Severity: af.Severity, After: af.Status}
		if ok {
			c.Before = bf.Status
		}
		out = append(out, c)
	}
	for id, bf := range b {
		if _, ok := a[id]; !ok {
			out = append(out, Change{ControlID: id, Title: bf.Title, Severity: bf.Severity, Before: bf.Status})
		}
	}

	rank := map[string]int{"regressed": 0, "improved": 1, "changed": 2}
	sort.Slice(out, func(i, j int) bool {
		di, dj := rank[out[i].Direction()], rank[out[j].Direction()]
		if di != dj {
			return di < dj
		}
		wi, wj := out[i].Severity.Weight(), out[j].Severity.Weight()
		if wi != wj {
			return wi > wj
		}
		return out[i].ControlID < out[j].ControlID
	})
	return out
}
