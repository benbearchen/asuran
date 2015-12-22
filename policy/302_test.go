package policy

import "testing"

func TestDont302Policy(t *testing.T) {
	cmd := "dont302"
	p, err := Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") failed: %v`, cmd, err)
	} else {
		if p.Command() != cmd {
			t.Errorf(`Factory("%s").Command() failed: "%s"`, cmd, p.Command())
		} else {
			p, ok := p.(*Dont302Policy)
			if !ok {
				t.Errorf(`Factory("%s") invalid class`, cmd)
			} else if !p.Value() {
				t.Errorf(`Factory("%s").Value() value unmatch %v`, cmd, p.Value())
			}
		}
	}

	cmd = "do302"
	p, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") failed: %v`, cmd, err)
	} else {
		if p.Command() != cmd {
			t.Errorf(`Factory("%s").Command() failed: "%s"`, cmd, p.Command())
		} else {
			p, ok := p.(*Dont302Policy)
			if !ok {
				t.Errorf(`Factory("%s") invalid class`, cmd)
			} else if p.Value() {
				t.Errorf(`Factory("%s").Value() value unmatch %v`, cmd, p.Value())
			}
		}
	}
}
