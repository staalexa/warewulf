package overlaydiff

import "sort"

// Decision records interactive selection for a path.
type Decision string

const (
	DecisionUnset     Decision = "unset"
	DecisionYes       Decision = "yes"
	DecisionNo        Decision = "no"
	DecisionTemplated Decision = "templated"
)

// ChangeSummary is persisted session metadata for a detected path.
type ChangeSummary struct {
	Path   string     `json:"path"`
	Change ChangeType `json:"change"`
	Type   EntryType  `json:"type"`
}

// SummarizeChanges converts detailed changes into a persistent summary.
func SummarizeChanges(changes []Change) []ChangeSummary {
	if len(changes) == 0 {
		return nil
	}

	result := make([]ChangeSummary, 0, len(changes))
	for _, change := range changes {
		result = append(result, ChangeSummary{Path: change.Path, Change: change.Change, Type: change.Type})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Path == result[j].Path {
			return result[i].Change < result[j].Change
		}
		return result[i].Path < result[j].Path
	})

	return result
}
