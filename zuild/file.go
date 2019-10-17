package zuild

import (
	"io/ioutil"

	"github.com/hashicorp/hcl/v2"

	"fmt"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

const (
	DefaultZfPath = "Zuildfile"
)

type ZuildFileInit struct {
	Help *Help  `hcl:"Help,block"`
	Args []*Arg `hcl:"Arg,block"`

	Remain hcl.Body `hcl:",remain"`

	hcl *hcl.File
}

type ZuildFile struct {
	Help   *Help    `hcl:"Help,block"`
	Args   []*Arg   `hcl:"Arg,block"`
	Tasks  []*Task  `hcl:"Task,block"`
	Remain hcl.Body `hcl:",remain"`

	taskIndex map[string]map[string]int
}

func ParseZuildFile(path string) (*ZuildFileInit, error) {
	var diag hcl.Diagnostics
	parser := hclparse.NewParser()

	hclBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	zf := &ZuildFileInit{}

	zf.hcl, diag = parser.ParseHCL(hclBytes, path)
	if diag.HasErrors() {
		return nil, diag
	}

	diag = gohcl.DecodeBody(zf.hcl.Body, nil, zf)
	if diag.HasErrors() {
		return nil, diag
	}

	return zf, err
}

func EvalZuildFile(zi *ZuildFileInit, ctx *hcl.EvalContext) (*ZuildFile, error) {
	zf := &ZuildFile{taskIndex: make(map[string]map[string]int)}

	diag := gohcl.DecodeBody(zi.hcl.Body, ctx, zf)
	if diag.HasErrors() {
		return nil, diag
	}

	// Index the document
	schema, _ := gohcl.ImpliedBodySchema(&ZuildFile{})
	content, _ := zi.hcl.Body.Content(schema)

	taskSchema, _ := gohcl.ImpliedBodySchema(&Task{})

	for _, task := range content.Blocks.OfType("Task") {

		taskContent, _ := task.Body.Content(taskSchema)
		zf.taskIndex[task.Labels[0]] = make(map[string]int)
		for _, sub := range taskContent.Blocks {
			zf.taskIndex[task.Labels[0]][fmt.Sprint(sub.Type, ".", sub.Labels[0])] = sub.TypeRange.Start.Line
		}
	}

	return zf, nil
}
