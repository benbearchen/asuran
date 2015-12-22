package policy

import "testing"

func TestDelayPolicy(t *testing.T) {
	cmd := "delay 1s"
	p, err := Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") failed: %v`, cmd, err)
	} else {
		if p.Command() != cmd {
			t.Errorf(`Factory("%s").Command() failed: "%s"`, cmd, p.Command())
		}

		d, ok := p.(*DelayPolicy)
		if !ok {
			t.Errorf(`Factory("%s") invalid class`, cmd)
		} else if d.duration != 1 {
			t.Errorf(`Factory("%s") duration wrong: %f`, cmd, d.duration)
		}
	}

	cmd = "timeout rand 1h"
	p, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") failed: %v`, cmd, err)
	} else {
		if p.Command() != cmd {
			t.Errorf(`Factory("%s").Command() failed: "%s"`, cmd, p.Command())
		}

		d, ok := p.(*TimeoutPolicy)
		if !ok {
			t.Errorf(`Factory("%s") invalid class`, cmd)
		} else {
			if d.duration != 3600 {
				t.Errorf(`Factory("%s") duration wrong: %f`, cmd, d.duration)
			}

			if !d.rand {
				t.Errorf(`Factory("%s") missing rand`, cmd)
			}
		}
	}

	cmd = "drop body 1ms"
	p, err = Factory(cmd)
	if err != nil {
		t.Errorf(`Factory("%s") failed: %v`, cmd, err)
	} else {
		if p.Command() != cmd {
			t.Errorf(`Factory("%s").Command() failed: "%s"`, cmd, p.Command())
		}

		d, ok := p.(*DropPolicy)
		if !ok {
			t.Errorf(`Factory("%s") invalid class`, cmd)
		} else {
			if d.duration != 0.001 {
				t.Errorf(`Factory("%s") duration wrong: %f`, cmd, d.duration)
			}

			if !d.body {
				t.Errorf(`Factory("%s") missing rand`, cmd)
			}
		}
	}
}
