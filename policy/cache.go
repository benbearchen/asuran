package policy

const cacheKeyword = "cache"

type CachePolicy struct {
	dogPolicy
}

func init() {
	regFactory(newDogPolicyFactory(cacheKeyword, func() Policy {
		return &CachePolicy{dogPolicy{cacheKeyword, "缓存"}}
	}))
}
