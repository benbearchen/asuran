package policy

import "testing"

func TestHostPolicy(t *testing.T) {
	cmd := "host 192.168.1.1"
	p, err := Factory(cmd)
	if err != nil {
		t.Errorf(`Factory(%s) meet err: %v`, cmd, err)
	} else if _, ok := p.(*HostPolicy); !ok {
		t.Errorf(`Factory(%s) wrong type: %v / %s`, cmd, p, p.Command())
	} else if p.(*HostPolicy).Host() != "192.168.1.1" {
		t.Errorf(`Factory(%s).Host() wrong: %s`, cmd, p.(*HostPolicy).Host())
	} else if p.(*HostPolicy).HTTP() != "192.168.1.1:80" {
		t.Errorf(`Factory(%s).HTTP() wrong(not 192.168.1.1:80): %s`, cmd, p.(*HostPolicy).HTTP())
	}

	cmd = "host 192.168.1.1:81"
	p, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory(%s) meet err: %v`, cmd, err)
	} else if _, ok := p.(*HostPolicy); !ok {
		t.Errorf(`Factory(%s) wrong type: %v / %s`, cmd, p, p.Command())
	} else if p.(*HostPolicy).Host() != "192.168.1.1:81" {
		t.Errorf(`Factory(%s).Host() wrong: %s`, cmd, p.(*HostPolicy).Host())
	} else if p.(*HostPolicy).HTTP() != "192.168.1.1:81" {
		t.Errorf(`Factory(%s).HTTP() wrong(not 192.168.1.1:81): %s`, cmd, p.(*HostPolicy).HTTP())
	}
}
