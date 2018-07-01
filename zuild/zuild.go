package zuild

import (
	"os"
	"sort"
	"strings"

	"context"
	"fmt"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/solvent-io/zkit/provider"
	"github.com/solvent-io/zuild/cli"
	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
)

const (
	DefaultTaskName = "default"
)

type Zuild struct {
	ui  *cli.Ui
	cmd *cobra.Command
	zf  *ZuildFile
	ctx *hcl.EvalContext
}

func New(ui *cli.Ui, cmd *cobra.Command, zi *ZuildFileInit) (*Zuild, error) {
	var err error

	z := &Zuild{}
	z.ui = ui
	z.cmd = cmd

	z.zf, err = z.eval(zi)
	if err != nil {
		return nil, err
	}

	return z, nil
}

func (z *Zuild) Run(task string) error {
	graph := NewTaskGraph()
	graph.Populate(z.zf.Tasks)

	options := z.options()

	tasks, err := graph.Get(z.taskOrDefault(task))
	if err != nil {
		return err
	}

	for index, task := range tasks {
		z.ui.Info(task.Name)

		for _, action := range task.Actions(z.zf.taskIndex[task.Name]) {
			if action.Condition() == nil || *action.Condition() == true {
				z.ui.Info(fmt.Sprint("* ", action.Type(), " [", action.Key(), "]"))

				ctx := context.WithValue(context.Background(), "options", options)
				prov := provider.Get(action)

				_, err := prov.Realize("build", ctx)
				if err != nil {
					z.ui.Fatal(err.Error())
				}
			} else {
				z.ui.Warn(fmt.Sprint("- ", action.Type(), " [", action.Key(), "]"))
			}
		}

		if index+1 != len(tasks) {
			z.ui.Out("")
		}
	}

	return nil
}

func (z *Zuild) List() error {
	if z.zf.Help.Title != "" {
		z.ui.Out(z.zf.Help.Title)
		z.ui.Out("")
	}

	z.ui.Out("Tasks:")

	for _, task := range z.zf.Tasks {
		z.ui.Out(task.Name)
	}

	return nil
}

func (z *Zuild) Graph(task string) error {
	if z.zf.Help.Title != "" {
		z.ui.Out(z.zf.Help.Title)
		z.ui.Out("")
	}

	z.ui.Out("Tasks:")

	graph := NewTaskGraph()
	graph.Populate(z.zf.Tasks)

	tasks, err := graph.Get(z.taskOrDefault(task))
	if err != nil {
		return err
	}

	for _, task := range tasks {
		z.ui.Out(task.Name)
	}

	return nil
}

func (z *Zuild) eval(zi *ZuildFileInit) (*ZuildFile, error) {
	z.ctx = &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	// Populate arg namespace
	args := make(map[string]cty.Value)
	for _, arg := range zi.Args {
		val, _ := z.cmd.Flags().GetString(arg.Name)
		args[arg.Name] = cty.StringVal(val)
	}
	z.ctx.Variables["arg"] = cty.ObjectVal(args)

	// Populate env namespace
	envs := make(map[string]cty.Value)
	for _, env := range os.Environ() {
		key := strings.Split(env, "=")[0]
		val, _ := os.LookupEnv(key)
		envs[key] = cty.StringVal(val)
	}
	z.ctx.Variables["env"] = cty.ObjectVal(envs)

	// Populate var namespace
	vars := make(map[string]cty.Value)
	attrs, _ := zi.Remain.JustAttributes()
	var attar []*hcl.Attribute
	for key := range attrs {
		attar = append(attar, attrs[key])
	}

	sort.Slice(attar, func(i, j int) bool {
		return attar[i].Range.Start.Line < attar[j].Range.Start.Line
	})

	for _, attr := range attar {
		val, diag := attr.Expr.Value(z.ctx)
		if diag.HasErrors() {
			return nil, diag
		}

		vars[attr.Name] = val
		z.ctx.Variables["var"] = cty.ObjectVal(vars)
	}

	return EvalZuildFile(zi, z.ctx)
}

func (z *Zuild) options() *provider.Options {
	verbose, _ := z.cmd.Flags().GetBool("Verbose")

	return &provider.Options{Verbose: verbose}
}

func (z *Zuild) taskOrDefault(task string) string {
	if task == "" {
		return DefaultTaskName
	}

	return task
}
