package policy

import "testing"

func TestSetPolicy(t *testing.T) {
	cmd := setKeyword
	s, err := Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") err: %v`, cmd, err)
	} else if s.Keyword() != setKeyword {
		t.Errorf(`Factory("%s").Keyword() not match %s: %s`, cmd, setKeyword, s.Keyword())
	} else if s.Command() != cmd {
		t.Errorf(`Factory("%s").Command() not match %s: %s`, cmd, cmd, s.Command())
	}

	cmd = updateKeyword
	s, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") err: %v`, cmd, err)
	} else if s.Keyword() != setKeyword {
		t.Errorf(`Factory("%s").Keyword() not match %s: %s`, cmd, setKeyword, s.Keyword())
	} else if s.Command() != cmd {
		t.Errorf(`Factory("%s").Command() not match %s: %s`, cmd, cmd, s.Command())
	}
}
