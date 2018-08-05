package zuild

import (
	"os"
	"sort"
	"strings"

	"context"
	"fmt"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/solvent-io/zkit/provider"
		"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/chuckpreslar/emission"
)

const (
	DefaultTaskName = "default"
)

type Zuild struct {
	cmd *cobra.Command
	zf  *ZuildFile
	ctx *hcl.EvalContext

	*emission.Emitter
}

func New(cmd *cobra.Command, zi *ZuildFileInit) (*Zuild, error) {
	var err error

	z := &Zuild{}
	z.cmd = cmd
	z.Emitter = emission.NewEmitter()

	z.ctx = &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

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
		z.Emit("task.header", task.Name)

		for _, action := range task.Actions(z.zf.taskIndex[task.Name]) {
			if action.Condition() == nil || *action.Condition() == true {
				z.Emit("action.header", fmt.Sprint(action.Type(), " [", action.Key(), "]"))

				ctx := context.WithValue(context.Background(), "options", options)
				ctx = context.WithValue(ctx, "phase", "build")
				prov := provider.Get(action, z.Emitter)

				err := prov.Realize(ctx)
				if err != nil && action.MayFail() == false {
					z.Emit("action.error", err.Error())
					z.fatal()
				} else if err != nil {
					z.Emit("action.error", err.Error())
				}
			} else {
				z.Emit("action.warn", fmt.Sprint(action.Type(), " [", action.Key(), "]"))
			}
		}

		if index+1 != len(tasks) {
			z.Emit("out", "")
		}
	}

	return nil
}

func (z *Zuild) List() error {
	if z.zf.Help.Title != "" {
		z.Emit("out", z.zf.Help.Title)
		z.Emit("out", "")
	}

	z.Emit("out", "Tasks:")

	for _, task := range z.zf.Tasks {
		z.Emit("out", task.Name)
	}

	return nil
}

func (z *Zuild) Graph(task string) error {
	if z.zf.Help.Title != "" {
		z.Emit("out", z.zf.Help.Title)
		z.Emit("out", "")
	}

	z.Emit("out", "Tasks:")

	graph := NewTaskGraph()
	graph.Populate(z.zf.Tasks)

	tasks, err := graph.Get(z.taskOrDefault(task))
	if err != nil {
		return err
	}

	for _, task := range tasks {
		z.Emit("out", task.Name)
	}

	return nil
}

func (z *Zuild) eval(zi *ZuildFileInit) (*ZuildFile, error) {
	// Add a test function
	z.ctx.Functions = map[string]function.Function{
		"fruit": function.New(&function.Spec{
			Type: function.StaticReturnType(cty.String),
			Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {

				return cty.StringVal("fruity!"), nil
			},
		},
		),
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

func (z *Zuild) options() *provider.ProviderOptions {
	verbose, _ := z.cmd.Flags().GetBool("Verbose")

	return &provider.ProviderOptions{Verbose: verbose}
}

func (z *Zuild) taskOrDefault(task string) string {
	if task == "" {
		return DefaultTaskName
	}

	return task
}

func (z *Zuild) cleanup() {

}

func (z *Zuild) fatal() {
	z.cleanup()
	os.Exit(1)
}
