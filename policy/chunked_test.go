package policy

import "testing"

func TestChunkedPolicy(t *testing.T) {
	cmd := `chunked default`
	p, err := Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") err: %v`, cmd, err)
	} else if p.Command() != cmd {
		t.Errorf(`Factory("%s").Command() wrong: %s`, cmd, p.Command())
	}

	cmd = `chunked on`
	p, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") err: %v`, cmd, err)
	} else if p.Command() != cmd {
		t.Errorf(`Factory("%s").Command() wrong: %s`, cmd, p.Command())
	}

	cmd = `chunked off`
	p, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") err: %v`, cmd, err)
	} else if p.Command() != cmd {
		t.Errorf(`Factory("%s").Command() wrong: %s`, cmd, p.Command())
	}

	cmd = `chunked block 3`
	p, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") err: %v`, cmd, err)
	} else if p.Command() != cmd {
		t.Errorf(`Factory("%s").Command() wrong: %s`, cmd, p.Command())
	}

	cmd = `chunked size 1,3,5,6`
	p, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") err: %v`, cmd, err)
	} else if p.Command() != cmd {
		t.Errorf(`Factory("%s").Command() wrong: %s`, cmd, p.Command())
	}

}
