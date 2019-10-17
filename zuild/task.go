package zuild

import (
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/sigil66/zkit/action"
	"gonum.org/v1/gonum/graph"
)

type Task struct {
	Name    string       `hcl:"name,label"`
	Require []string     `hcl:"require,optional"`
	Sh      []*action.Sh `hcl:"Sh,block"`

	Config hcl.Body `hcl:",remain"`

	Node graph.Node
}

func (t *Task) Actions(index map[string]int) action.Actions {
	var actions action.Actions

	for index := range t.Sh {
		actions = append(actions, t.Sh[index])
	}

	sort.Slice(actions, func(i, j int) bool {
		return index[actions[i].Key()] < index[actions[j].Key()]
	})

	return actions
}
