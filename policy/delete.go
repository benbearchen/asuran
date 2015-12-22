package policy

const deleteKeyword = "delete"

type DeletePolicy struct {
	dogPolicy
}

func init() {
	regFactory(newDogPolicyFactory(deleteKeyword, func() Policy {
		return &DeletePolicy{dogPolicy{deleteKeyword, "删除"}}
	}))
}
