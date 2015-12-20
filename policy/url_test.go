package policy

import "testing"

func TestUrlPolicy(t *testing.T) {
	cmd := "url speed 1B/s g.cn"
	u, err := Factory(cmd)
	if err != nil {
		t.Errorf("url(%s) failed: %v", cmd, err)
	} else {
		if u.Command() != cmd {
			t.Errorf("url(%s).Command() changed: %s", cmd, u.Command())
		}

		url, ok := u.(*UrlPolicy)
		if !ok {
			t.Errorf("url(%s) result not *UrlPolicy", cmd, u)
		} else {
			if url.Target() != "g.cn" {
				t.Errorf(`url(%s).Target() wrong: %s`, cmd, url.Target())
			}

			if url.Speed() == nil {
				t.Errorf("url(%s) missed speed policy", cmd)
			}

			if url.Status() != nil {
				t.Errorf("url(%s) should not have status policy", cmd)
			}
		}
	}

	cmd = "url drop 1s status 400 g.cn"
	u, err = Factory(cmd)
	if err != nil {
		t.Errorf("url(%s) failed: %v", cmd, err)
	} else {
		if u.Command() != cmd {
			t.Errorf("url(%s).Command() changed: %s", cmd, u.Command())
		}

		url, ok := u.(*UrlPolicy)
		if !ok {
			t.Errorf("url(%s) result not *UrlPolicy", cmd, u)
		} else {
			if url.DelayPolicy() == nil {
				t.Errorf("url(%s) missed drop policy", cmd)
			} else {
				d, ok := url.DelayPolicy().(*DropPolicy)
				if !ok {
					t.Errorf("url(%s).DelayPolicy() not drop policy: %v", cmd, url.DelayPolicy())
				} else if d.duration != 1 {
					t.Errorf("url(%s).DelayPolicy().duration is wrong: %d", cmd, d.duration)
				}
			}

			p := url.Status()
			if p == nil {
				t.Errorf("url(%s) missed status policy", cmd)
			} else if p.StatusCode() != 400 {
				t.Errorf("url(%s).Status().StatusCode() is wrong: %d", cmd, p.StatusCode())
			}
		}
	}

	cmdex := "url status 502 g.cn"
	ex, err := Factory(cmdex)
	if err != nil {
		t.Errorf("url(%s) failed: %v", cmdex, err)
	} else {
		err = u.Update(ex)
		if err != nil {
			t.Errorf("url(%s).Update(%s) failed: %v", cmd, cmdex, err)
		}

		if u.Command() != "url drop 1s status 502 g.cn" {
			t.Errorf("url(%s).Update(%s).Command() changed: %s", cmd, cmdex, u.Command())
		}
	}

	cmd = u.Command()
	cmdex = "url update speed 1"
	ex, err = Factory(cmdex)
	if err != nil {
		t.Errorf("url(%s) failed: %v", cmdex, err)
	} else {
		if ex.(*UrlPolicy).Set() {
			t.Errorf("url(%s).Set() not false", cmdex)
		}

		err = u.Update(ex)
		if err != nil {
			t.Errorf("url(%s).Update(%s) failed: %v", cmd, cmdex, err)
		}

		p := u.(*UrlPolicy)
		if p.Speed() == nil || p.Speed().Speed() != 1 {
			t.Errorf("url(%s).Update(%s).Speed() wrong: %v", cmd, cmdex, p.Speed())
		}

		if p.Status() == nil || p.Status().StatusCode() != 502 {
			t.Errorf("url(%s).Update(%s).Status() wrong: %v", cmd, cmdex, p.Status())
		}
	}

	cmd = u.Command()
	cmdex = "url set dont302"
	ex, err = Factory(cmdex)
	if err != nil {
		t.Errorf("url(%s) failed: %v", cmdex, err)
	} else {
		if !ex.(*UrlPolicy).Set() {
			t.Errorf("url(%s).Set() not true", cmdex)
		}

		err = u.Update(ex)
		if err != nil {
			t.Errorf("url(%s).Update(%s) failed: %v", cmd, cmdex, err)
		}

		p := u.(*UrlPolicy)
		if p.Speed() != nil {
			t.Errorf("url(%s).Update(%s).Speed() not be nil: %v", cmd, cmdex, p.Speed())
		}

		if p.Status() != nil {
			t.Errorf("url(%s).Update(%s).Status() not be nil: %v", cmd, cmdex, p.Status())
		}

		if p.Dont302() == false {
			t.Errorf("url(%s).Update(%s).Dont302() not be true", cmd, cmdex)
		}

		if p.DelayPolicy() != nil {
			t.Errorf("url(%s).Update(%s).DelayPolicy() not be nil: %v", cmd, cmdex, p.DelayPolicy())
		}
	}
}
