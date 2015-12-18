package policy

const rewriteKeyword = "rewrite"

type RewritePolicy struct {
	stringPolicy
}

func init() {
	regFactory(newStringPolicyFactory(rewriteKeyword, "url encoded content", func(content string) (Policy, error) {
		err := checkEncodedContent(content)
		if err != nil {
			return nil, err
		} else {
			return &RewritePolicy{stringPolicy{rewriteKeyword, content, func(string) string {
				return "以特定内容返回"
			}}}, nil
		}
	}))
}

func (r *RewritePolicy) Content() ([]byte, error) {
	return decodeContent(r.str)
}
