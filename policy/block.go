package policy

const blockKeyword = "block"

type BlockPolicy struct {
	dogPolicy
}

func init() {
	regFactory(newDogPolicyFactory(blockKeyword, func() Policy {
		return &BlockPolicy{dogPolicy{blockKeyword, "丢弃"}}
	}))
}
