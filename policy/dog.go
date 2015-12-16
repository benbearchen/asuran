package policy

type dogPolicyFactory struct {
	keyword string
	create  func() Policy
}

func newDogPolicyFactory(keyword string, create func() Policy) *dogPolicyFactory {
	return &dogPolicyFactory{keyword, create}
}

func (dog *dogPolicyFactory) Keyword() string {
	return dog.keyword
}

func (dog *dogPolicyFactory) Build(args []string) (Policy, []string, error) {
	return dog.create(), args, nil
}

type dogPolicy struct {
	keyword string
	comment string
}

func (dog *dogPolicy) Keyword() string {
	return dog.keyword
}

func (dog *dogPolicy) Command() string {
	return dog.keyword
}

func (dog *dogPolicy) Comment() string {
	return dog.comment
}
