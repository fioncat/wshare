package clipboard

import (
	"sync"
	"time"

	"github.com/fioncat/wshare/pkg/osutil"
)

var (
	imageCooldown = newCooldown()
	textCooldown  = newCooldown()
)

const cooldownSeconds int64 = 10

type cooldownSet struct {
	lock sync.Mutex

	data map[string]int64
}

func newCooldown() *cooldownSet {
	return &cooldownSet{data: make(map[string]int64)}
}

func (s *cooldownSet) Set(data []byte) {
	key := osutil.Sum(data)
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data[key] = time.Now().Unix() + cooldownSeconds
}

func (s *cooldownSet) Exists(data []byte) bool {
	key := osutil.Sum(data)
	s.lock.Lock()
	defer s.lock.Unlock()
	expired, exists := s.data[key]
	if !exists {
		return false
	}
	now := time.Now().Unix()
	if now >= expired {
		delete(s.data, key)
		return false
	}
	return true
}

func (s *cooldownSet) cleanup() {
	tk := time.Tick(time.Minute)
	for range tk {
		s.lock.Lock()
		now := time.Now().Unix()
		for key, expired := range s.data {
			if now >= expired {
				delete(s.data, key)
			}
		}
		s.lock.Unlock()
	}
}
