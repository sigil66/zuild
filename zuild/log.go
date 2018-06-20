package zuild

import (
"fmt"

)

type Log struct {
	Name string `hcl:"name,label"`
	Message string `hcl:"message"`
}

func (l *Log) Id() string {
	return l.Name
}

func (l *Log) Key() string {
	return fmt.Sprint(l.Type(), ".", l.Name)
}

func (l *Log) Realize() (string, error) {
	return l.Message, nil
}

func (l *Log) Type() string {
	return "Log"
}