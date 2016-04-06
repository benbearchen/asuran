package policy

import "testing"

func TestDeletePolicy(t *testing.T) {
	cmd := "url delete g.cn"
	p, err := Factory(cmd)
	if err != nil {
		t.Errorf(`Factory(%s) err: %v`, cmd, err)
	} else {
		p := p.(*UrlPolicy)
		if p.Delete() != true {
			t.Errorf(`Factory(%s).Delete() not true`, cmd)
		}
	}
}
