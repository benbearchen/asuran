package policy

import "testing"

func TestDomainPolicy(t *testing.T) {
	cmd := "domain xyz"
	d, err := Factory(cmd)
	if err != nil {
		t.Errorf("domain(%s) failed: ", cmd, err)
	} else if d.Command() != cmd {
		t.Errorf("domain(%s).Command() changed: %s", cmd, d.Command())
	}

	cmd = "domain xyz abc"
	d, err = Factory(cmd)
	if err == nil {
		t.Errorf("domain(%s) didn't detect error", cmd)
	}

	cmd = "domain block"
	d, err = Factory(cmd)
	if err == nil {
		t.Errorf("domain(%s) didn't detect error", cmd)
	}

	cmd = "domain block g.cn"
	d, err = Factory(cmd)
	if err != nil {
		t.Errorf("domain(%s) failed: ", cmd, err)
	} else if d.Command() != cmd {
		t.Errorf("domain(%s).Command() changed: %s", cmd, d.Command())
	} else if _, ok := d.(*DomainPolicy); !ok {
		t.Errorf("domain(%s) return non *DomainPolicy", cmd, d)
	} else {
		d := d.(*DomainPolicy)
		if _, ok := d.Action().(*BlockPolicy); !ok {
			t.Errorf("domain(%s).Action() is not *BlockPolicy", cmd, d.Action())
		}
	}
}
