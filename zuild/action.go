package zuild

type Action interface {
	Id() string
	Key() string
	Realize() (string, error)
	Type() string
}

type Actions []Action


