package policy

const restoreKeyword = "restore"

type RestorePolicy struct {
	stringPolicy
}

func init() {
	regFactory(newStringPolicyFactory(restoreKeyword, "stored id", func(id string) (Policy, error) {
		return &RestorePolicy{stringPolicy{restoreKeyword, id, func(id string) string {
			return "以预定义 " + id + " 内容返回"
		}}}, nil
	}))
}
