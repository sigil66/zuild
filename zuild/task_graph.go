package zuild

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
	"sort"
)

type TaskGraph struct {
	graph     *simple.DirectedGraph
	root      graph.Node

	docIndex  map[graph.Node]int
	nodeIndex map[graph.Node]*Task
	nameIndex map[string]*Task
}

func NewTaskGraph() *TaskGraph {
	tg := &TaskGraph{}

	tg.graph = simple.NewDirectedGraph()
	tg.root = tg.graph.NewNode()

	tg.docIndex = make(map[graph.Node]int)
	tg.nodeIndex = make(map[graph.Node]*Task)
	tg.nameIndex = make(map[string]*Task)

	return tg
}

func (t *TaskGraph) Populate(tasks []*Task) {
	for index := range tasks {
		t.addNode(tasks[index])
	}

	for index := range tasks {
		t.addEdges(tasks[index])
	}
}

func (t *TaskGraph) Get(root string) ([]*Task, error) {
	var tasks []*Task

	nodes, err := topo.SortStabilized(t.graph, t.sortStable)
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if topo.PathExistsIn(t.graph, node, t.nameIndex[root].Node) {
			tasks = append(tasks, t.nodeIndex[node])
		}
	}

	return tasks, nil
}

func (t *TaskGraph) addNode(task *Task) {
	task.Node = t.graph.NewNode()
	t.graph.AddNode(task.Node)

	t.nodeIndex[task.Node] = task
	t.nameIndex[task.Name] = task
}

func (t *TaskGraph) addEdges(task *Task) {
	for _, req := range task.Require {
		edge := t.graph.NewEdge(t.nameIndex[req].Node, task.Node)
		t.graph.SetEdge(edge)

		t.docIndex[t.nameIndex[req].Node] = len(t.docIndex)
	}
}

func (t *TaskGraph) sortStable(nodes []graph.Node) {
	sort.SliceStable(nodes, func(i, j int) bool {
		return t.docIndex[nodes[i]] <  t.docIndex[nodes[j]]
	})
}
