package policy

const proxyKeyword = "proxy"

type ProxyPolicy struct {
	dogPolicy
}

func init() {
	regFactory(newDogPolicyFactory(proxyKeyword, func() Policy {
		return &ProxyPolicy{dogPolicy{proxyKeyword, "代理"}}
	}))
}
