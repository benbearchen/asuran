package policy

const nullKeyword = "null"

type NullPolicy struct {
	dogPolicy
}

func init() {
	regFactory(newDogPolicyFactory(nullKeyword, func() Policy {
		return &NullPolicy{dogPolicy{nullKeyword, "查询无结果"}}
	}))
}
