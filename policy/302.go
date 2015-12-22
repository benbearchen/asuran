package policy

const dont302Keyword = "dont302"
const do302Keyword = "do302"

type Dont302Policy struct {
	boolPolicy
}

func init() {
	regFactory(newBoolPolicyFactory(dont302Keyword, func() Policy {
		return &Dont302Policy{boolPolicy{dont302Keyword, true, "允许 302 穿透"}}
	}))

	regFactory(newBoolPolicyFactory(do302Keyword, func() Policy {
		return &Dont302Policy{boolPolicy{do302Keyword, false, "捕获 302 跳转"}}
	}))
}
