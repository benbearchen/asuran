package policy

const defaultKeyword = "default"

type DefaultPolicy struct {
	dogPolicy
}

func init() {
	regFactory(newDogPolicyFactory(defaultKeyword, func() Policy {
		return &DefaultPolicy{dogPolicy{defaultKeyword, "默认"}}
	}))
}
