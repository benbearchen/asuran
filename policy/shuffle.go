package policy

const shuffleKeyword = "shuffle"

type ShufflePolicy struct {
	dogPolicy
}

func init() {
	regFactory(newDogPolicyFactory(shuffleKeyword, func() Policy {
		return &ShufflePolicy{dogPolicy{shuffleKeyword, "乱序"}}
	}))
}
