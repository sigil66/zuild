package zuild

type Arg struct {
	Name  string `hcl:"name,label"`
	Short string `hcl:"short"`
	Usage string `hcl:"usage"`
}
