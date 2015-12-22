package policy

const disable304Keyword = "disable304"
const allow304Keyword = "allow304"

type Disable304Policy struct {
	boolPolicy
}

func init() {
	regFactory(newBoolPolicyFactory(disable304Keyword, func() Policy {
		return &Disable304Policy{boolPolicy{disable304Keyword, true, "禁止 304"}}
	}))

	regFactory(newBoolPolicyFactory(allow304Keyword, func() Policy {
		return &Disable304Policy{boolPolicy{allow304Keyword, false, "允许 304"}}
	}))
}
