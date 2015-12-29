package policy

import "testing"

func TestToolReplacer(t *testing.T) {
	check := func(r *Replacer, pattern, source, target string) {
		result := r.Replace(source)
		if result != target {
			t.Errorf(`NewReplacer("%s").Replace(%s)=> %s != %s`, pattern, source, result, target)
		}
	}

	pattern := "/ab/cd/"
	r, err := NewReplacer(pattern)
	if err != nil {
		t.Errorf(`NewReplacer("%s") err: `, pattern, err)
	} else {
		check(r, pattern, "", "")
		check(r, pattern, "a", "a")
		check(r, pattern, "ab", "cd")
	}

	pattern = "/ab/$1/"
	r, err = NewReplacer(pattern)
	if err != nil {
		t.Errorf(`NewReplacer("%s") err: `, pattern, err)
	} else {
		check(r, pattern, "", "")
		check(r, pattern, "a", "a")
		check(r, pattern, "ab", "")
	}

	pattern = "/ab/x${1}y/"
	r, err = NewReplacer(pattern)
	if err != nil {
		t.Errorf(`NewReplacer("%s") err: `, pattern, err)
	} else {
		check(r, pattern, "", "")
		check(r, pattern, "a", "a")
		check(r, pattern, "ab", "xy")
	}

	pattern = "/(ab)/x${1}y/"
	r, err = NewReplacer(pattern)
	if err != nil {
		t.Errorf(`NewReplacer("%s") err: `, pattern, err)
	} else {
		check(r, pattern, "", "")
		check(r, pattern, "a", "a")
		check(r, pattern, "ab", "xaby")
	}
}
