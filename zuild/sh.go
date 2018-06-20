package zuild

import (
	"fmt"
	"strings"
)

type Sh struct {
	Name string `hcl:"name,label"`
	Cmd []string `hcl:"cmd"`
}

func (s *Sh) Id() string {
	return s.Name
}

func (s *Sh) Key() string {
	return fmt.Sprint(s.Type(), ".", s.Name)
}

func (s *Sh) Realize() (string, error) {
	return strings.Join(s.Cmd, " "), nil
}

func (s *Sh) Type() string {
	return "Sh"
}