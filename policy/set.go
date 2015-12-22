package policy

const setKeyword = "set"
const updateKeyword = "update"

type SetPolicy struct {
	boolPolicy
}

func init() {
	regFactory(newBoolPolicyFactory(setKeyword, func() Policy {
		return &SetPolicy{boolPolicy{setKeyword, true, "覆盖设置"}}
	}))

	regFactory(newBoolPolicyFactory(updateKeyword, func() Policy {
		return &SetPolicy{boolPolicy{updateKeyword, false, "更新"}}
	}))
}
