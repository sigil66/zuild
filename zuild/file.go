package zuild

import (
	"io/ioutil"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
)

const (
	DefaultZfPath = "Zuildfile"
)

type ZuildFileInit struct {
	Help   *Help    `hcl:"Help,block"`
	Args   []*Arg   `hcl:"Arg,block"`

	Remain hcl.Body `hcl:",remain"`

	hcl *hcl.File
}

type ZuildFile struct {
	Help   *Help    `hcl:"Help,block"`
	Args   []*Arg   `hcl:"Arg,block"`
	Tasks  []*Task  `hcl:"Task,block"`
	Remain hcl.Body `hcl:",remain"`
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
	zf := &ZuildFile{}

	diag := gohcl.DecodeBody(zi.hcl.Body, ctx, zf)
	if diag.HasErrors() {
		return nil, diag
	}

	return zf, nil
}
