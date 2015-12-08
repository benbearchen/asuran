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
		}
	}
}
