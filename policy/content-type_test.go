package policy

import "testing"

func TestContentTypePolicy(t *testing.T) {
	cmd := "contentType default"
	p, err := Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") failed: %v`, cmd, err)
	} else {
		if p.Command() != cmd {
			t.Errorf(`Factory("%s").Command() failed: "%s"`, cmd, p.Command())
		} else {
			p, ok := p.(*ContentTypePolicy)
			if !ok {
				t.Errorf(`Factory("%s") invalid class`, cmd)
			} else if p.Value() != ContentTypeActDefault {
				t.Errorf(`Factory("%s").Value() value unmatch %v`, cmd, p.Value())
			}
		}
	}

	cmd = "contentType remove"
	p, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") failed: %v`, cmd, err)
	} else {
		if p.Command() != cmd {
			t.Errorf(`Factory("%s").Command() failed: "%s"`, cmd, p.Command())
		} else {
			p, ok := p.(*ContentTypePolicy)
			if !ok {
				t.Errorf(`Factory("%s") invalid class`, cmd)
			} else if p.Value() != ContentTypeActRemove {
				t.Errorf(`Factory("%s").Value() value unmatch %v`, cmd, p.Value())
			}
		}
	}

	cmd = "contentType empty"
	p, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") failed: %v`, cmd, err)
	} else {
		if p.Command() != cmd {
			t.Errorf(`Factory("%s").Command() failed: "%s"`, cmd, p.Command())
		} else {
			p, ok := p.(*ContentTypePolicy)
			if !ok {
				t.Errorf(`Factory("%s") invalid class`, cmd)
			} else if p.Value() != ContentTypeActEmpty {
				t.Errorf(`Factory("%s").Value() value unmatch %v`, cmd, p.Value())
			}
		}
	}

	cmd = "contentType text/html"
	p, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") failed: %v`, cmd, err)
	} else {
		if p.Command() != cmd {
			t.Errorf(`Factory("%s").Command() failed: "%s"`, cmd, p.Command())
		} else {
			p, ok := p.(*ContentTypePolicy)
			if !ok {
				t.Errorf(`Factory("%s") invalid class`, cmd)
			} else if p.Value() != "text/html" {
				t.Errorf(`Factory("%s").Value() value unmatch %v`, cmd, p.Value())
			}
		}
	}
}
