package policy

import "testing"

func TestDisable304Policy(t *testing.T) {
	cmd := disable304Keyword
	s, err := Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") err: %v`, cmd, err)
	} else if s.Keyword() != disable304Keyword {
		t.Errorf(`Factory("%s").Keyword() not match %s: %s`, cmd, disable304Keyword, s.Keyword())
	} else if s.Command() != cmd {
		t.Errorf(`Factory("%s").Command() not match %s: %s`, cmd, cmd, s.Command())
	}

	cmd = allow304Keyword
	s, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") err: %v`, cmd, err)
	} else if s.Keyword() != disable304Keyword {
		t.Errorf(`Factory("%s").Keyword() not match %s: %s`, cmd, disable304Keyword, s.Keyword())
	} else if s.Command() != cmd {
		t.Errorf(`Factory("%s").Command() not match %s: %s`, cmd, cmd, s.Command())
	}
}
