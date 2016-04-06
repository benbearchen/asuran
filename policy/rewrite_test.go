package policy

import "testing"

func TestRewritePolicy(t *testing.T) {
	cmd := "rewrite %20"
	p, err := Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") failed: %v`, cmd, err)
	} else {
		if p.Command() != cmd {
			t.Errorf(`Factory("%s").Command() failed: "%s"`, cmd, p.Command())
		} else {
			p, ok := p.(*RewritePolicy)
			if !ok {
				t.Errorf(`Factory("%s") invalid class`, cmd)
			} else if p.Value() != "%20" {
				t.Errorf(`Factory("%s").Value() value unmatch %v`, cmd, p.Value())
			} else {
				c, err := p.Content()
				if err != nil {
					t.Errorf(`Factory("%s").Content() failed: %v`, cmd, err)
				}

				if len(c) != 1 || c[0] != ' ' {
					t.Errorf(`Factory("%s").Content() value unmatch %v`, cmd, c)
				}
			}
		}
	}

}
