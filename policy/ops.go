package policy

const restartKeyword = "restart"

type RestartPolicy struct {
	dogPolicy
}

const clearKeyword = "clear"

type ClearPolicy struct {
	dogPolicy
}

func init() {
	regFactory(newDogPolicyFactory(restartKeyword, func() Policy {
		return &RestartPolicy{dogPolicy{restartKeyword, "重启历史"}}
	}))

	regFactory(newDogPolicyFactory(clearKeyword, func() Policy {
		return &ClearPolicy{dogPolicy{clearKeyword, "清空配置"}}
	}))
}
