package policy

const circularKeyword = "circular"

type CircularPolicy struct {
	dogPolicy
}

func init() {
	regFactory(newDogPolicyFactory(circularKeyword, func() Policy {
		return &CircularPolicy{dogPolicy{circularKeyword, "循环访问"}}
	}))
}
