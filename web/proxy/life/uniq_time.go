package life

import (
	"sync"
	"time"
)

var (
	lockUniqueTime sync.Mutex
	lastUniqueTime int64
)

func uniqueTime() time.Time {
	t := time.Now().UnixNano()

	lockUniqueTime.Lock()
	defer lockUniqueTime.Unlock()

	if lastUniqueTime < t {
		lastUniqueTime = t
	} else {
		lastUniqueTime++
		t = lastUniqueTime
	}

	return time.Unix(0, t)
}
