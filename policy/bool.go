package policy

type boolPolicyFactory struct {
	keyword string
	create  func() Policy
}

func newBoolPolicyFactory(keyword string, create func() Policy) *boolPolicyFactory {
	return &boolPolicyFactory{keyword, create}
}

func (b *boolPolicyFactory) Keyword() string {
	return b.keyword
}

func (b *boolPolicyFactory) Build(args []string) (Policy, []string, error) {
	return b.create(), args, nil
}

type boolPolicy struct {
	keyword string
	boolean bool
	comment string
}

func (b *boolPolicy) Keyword() string {
	return b.keyword
}

func (b *boolPolicy) Command() string {
	return b.keyword
}

func (b *boolPolicy) Comment() string {
	return b.comment
}

func (b *boolPolicy) Value() bool {
	return b.boolean
}
