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
			if url.Speed() == nil {
				t.Errorf("url(%s) missed speed policy", cmd)
			}

			if url.Status() != nil {
				t.Errorf("url(%s) should not have status policy", cmd)
			}
		}
	}

	cmd = "url status 400 g.cn"
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
			p := url.Status()
			if p == nil {
				t.Errorf("url(%s) missed status policy", cmd)
			} else if p.StatusCode() != 400 {
				t.Errorf("url(%s).Status().StatusCode() is wrong: %d", cmd, p.StatusCode())
			}
		}
	}
}
