package policy

const contentTypeKeyword = "contentType"

const (
	ContentTypeActDefault = "default"
	ContentTypeActRemove  = "remove"
	ContentTypeActEmpty   = "empty"
)

type ContentTypePolicy struct {
	stringPolicy
}

func init() {
	regFactory(newStringPolicyFactory(contentTypeKeyword, "operator or value", func(val string) (Policy, error) {
		return &ContentTypePolicy{stringPolicy{contentTypeKeyword, val, func(val string) string {
			switch val {
			case ContentTypeActDefault:
				return ""
			case ContentTypeActRemove:
				return "移除 Content-Type"
			case ContentTypeActEmpty:
				return "置空 Content-Type"
			default:
				return "Content-Type: " + val
			}
		}}}, nil
	}))
}
