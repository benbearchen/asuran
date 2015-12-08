package policy

type Policy interface {
	Keyword() string
	Command() string
	Comment() string
}
