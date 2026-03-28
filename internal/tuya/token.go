package tuya

import (
	"sync"
	"time"
)

type TokenProvider struct {
	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func (tp *TokenProvider) Get() (string, bool) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	if tp.token == "" || time.Until(tp.expiresAt) < 60*time.Second {
		return "", false
	}
	return tp.token, true
}

func (tp *TokenProvider) Set(token string, expireSeconds int) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.token = token
	tp.expiresAt = time.Now().Add(time.Duration(expireSeconds) * time.Second)
}

func (tp *TokenProvider) Invalidate() {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.token = ""
	tp.expiresAt = time.Time{}
}
