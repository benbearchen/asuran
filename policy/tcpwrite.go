package policy

const tcpwriteKeyword = "tcpwrite"

type TcpwritePolicy struct {
	stringPolicy
}

func init() {
	regFactory(newStringPolicyFactory(tcpwriteKeyword, "url encoded content", func(content string) (Policy, error) {
		err := checkEncodedContent(content)
		if err != nil {
			return nil, err
		} else {
			return &TcpwritePolicy{stringPolicy{tcpwriteKeyword, content, func(string) string {
				return "将特定内容从 TCP 返回"
			}}}, nil
		}
	}))
}

func (r *TcpwritePolicy) Content() ([]byte, error) {
	return decodeContent(r.str)
}
