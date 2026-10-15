package handlers

import (
	"strings"
	"sync"
	"time"
)

// loginLimiter slows password guessing by counting failed sign-ins per account
// from one address, per account overall and per address overall. Only
// failures count, and the counts live in memory, which suits one API instance.
type loginLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	limits  map[string]int
	entries map[string]*loginWindow
}

type loginWindow struct {
	failures int
	started  time.Time
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{
		window: 15 * time.Minute,
		// One account from one address is the brute-force case; the wider limits
		// stop spreading guesses over addresses or accounts without locking out
		// an office that shares one internet address.
		limits:  map[string]int{"account+ip": 5, "account": 20, "ip": 50},
		entries: map[string]*loginWindow{},
	}
}

func (l *loginLimiter) keys(email, ip string) map[string]string {
	account := strings.ToLower(strings.TrimSpace(email))
	return map[string]string{"account+ip": "a+i:" + account + "|" + ip, "account": "a:" + account, "ip": "i:" + ip}
}

// blocked reports how long the caller must wait, or zero when it may try.
func (l *loginLimiter) blocked(email, ip string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	var wait time.Duration
	for kind, key := range l.keys(email, ip) {
		entry := l.current(key, now)
		if entry != nil && entry.failures >= l.limits[kind] {
			if remaining := entry.started.Add(l.window).Sub(now); remaining > wait {
				wait = remaining
			}
		}
	}
	return wait
}

func (l *loginLimiter) failed(email, ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if len(l.entries) > 10000 {
		for key, entry := range l.entries {
			if now.Sub(entry.started) >= l.window {
				delete(l.entries, key)
			}
		}
	}
	for _, key := range l.keys(email, ip) {
		entry := l.current(key, now)
		if entry == nil {
			entry = &loginWindow{started: now}
			l.entries[key] = entry
		}
		entry.failures++
	}
}

// succeeded clears the account's failures; the address count stays so a
// valid account cannot be used to reset guessing against other accounts.
func (l *loginLimiter) succeeded(email, ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	keys := l.keys(email, ip)
	delete(l.entries, keys["account+ip"])
	delete(l.entries, keys["account"])
}

func (l *loginLimiter) current(key string, now time.Time) *loginWindow {
	entry := l.entries[key]
	if entry != nil && now.Sub(entry.started) >= l.window {
		delete(l.entries, key)
		return nil
	}
	return entry
}
