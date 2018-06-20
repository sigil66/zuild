package zuild

import (
	"github.com/hashicorp/hcl2/hcl"
	"gonum.org/v1/gonum/graph"
	"sort"
)

type Task struct {
	Name    string   `hcl:"name,label"`
	Require []string `hcl:"require,optional"`
	Sh []*Sh `hcl:"Sh,block"`
	Log []*Log `hcl:"Log,block"`

	Config hcl.Body `hcl:",remain"`

	Node graph.Node
}

func (t *Task) Actions(index map[string]int) Actions {
	var actions Actions
	
	for index := range t.Sh {
		actions = append(actions, t.Sh[index])
	}

	for index := range t.Log {
		actions = append(actions, t.Log[index])
	}
	
	sort.Slice(actions, func(i, j int) bool {
		return index[actions[i].Key()] < index[actions[j].Key()]
	})

	return actions
}