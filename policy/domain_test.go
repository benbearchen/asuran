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

	cmd = "domain block proxy abc"
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

func TestDomainPolicyOpts(t *testing.T) {
	cmd := "domain n 1 g.cn 192.168.1.1,192.168.1.2"
	d, err := Factory(cmd)
	if err != nil {
		t.Errorf("domain(%s) failed: ", cmd, err)
	} else if d.Command() != cmd {
		t.Errorf("domain(%s).Command() changed: %s", cmd, d.Command())
	}

	cmd = "domain g.cn shuffle"
	d, err = Factory(cmd)
	if err == nil {
		t.Errorf("domain(%s) didn't detect error", cmd)
	}

	cmd = "domain shuffle g.cn 192.168.1.1,192.168.1.2"
	d, err = Factory(cmd)
	if err != nil {
		t.Errorf("domain(%s) failed: ", cmd, err)
	} else if d.Command() != cmd {
		t.Errorf("domain(%s).Command() changed: %s", cmd, d.Command())
	} else {
		d := d.(*DomainPolicy)
		found := false
		for i := 0; i < 100; i++ {
			ips := d.NextIPs()
			if len(ips) == 2 && ips[0] == "192.168.1.2" && ips[1] == "192.168.1.1" {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("domain(%s).NextIPs() maybe not shuffled", cmd)
		}
	}

	cmd = "domain n -1 g.cn"
	d, err = Factory(cmd)
	if err == nil {
		t.Errorf("domain(%s) didn't detect error", cmd)
	}

	cmd = "domain n 0 g.cn 192.168.1.1,192.168.1.2"
	d, err = Factory(cmd)
	if err != nil {
		t.Errorf("domain(%s) failed: ", cmd, err)
	} else if d.Command() != cmd {
		t.Errorf("domain(%s).Command() changed: %s", cmd, d.Command())
	} else {
		d := d.(*DomainPolicy)
		ips := d.NextIPs()
		if len(ips) != 0 {
			t.Errorf("domain(%s).NextIPs() return not empty: %v", cmd, ips)
		}
	}

	cmd = "domain n 1 g.cn 192.168.1.1,192.168.1.2"
	d, err = Factory(cmd)
	if err != nil {
		t.Errorf("domain(%s) failed: ", cmd, err)
	} else if d.Command() != cmd {
		t.Errorf("domain(%s).Command() changed: %s", cmd, d.Command())
	} else {
		d := d.(*DomainPolicy)
		ips := d.NextIPs()
		if len(ips) != 1 {
			t.Errorf("domain(%s).NextIPs() return not one ip: %v", cmd, ips)
		}
	}

	cmd = "domain n 1 circular g.cn 192.168.1.1,192.168.1.2"
	d, err = Factory(cmd)
	if err != nil {
		t.Errorf("domain(%s) failed: ", cmd, err)
	} else {
		d := d.(*DomainPolicy)
		ips := d.NextIPs()
		if len(ips) != 1 {
			t.Errorf("domain(%s).NextIPs() return not one ip: %v", cmd, ips)
		} else if ips[0] != "192.168.1.1" {
			t.Errorf("domain(%s).NextIPs() return not [192.168.1.1]: %v", cmd, ips)
		}

		ips = d.NextIPs()
		if len(ips) != 1 || ips[0] != "192.168.1.2" {
			t.Errorf("domain(%s).NextIPs() return not [192.168.1.2]: %v", cmd, ips)
		}

		ips = d.NextIPs()
		if len(ips) != 1 || ips[0] != "192.168.1.1" {
			t.Errorf("domain(%s).NextIPs() return not [192.168.1.1]: %v", cmd, ips)
		}
	}
}
