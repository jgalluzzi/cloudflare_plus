// cloudflareplus/tokenrotator.go
package cloudflareplus

import (
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type TokenRotator struct {
	Tokens []string
	index  int
	lock   sync.Mutex
}

func (t *TokenRotator) RoundTrip(req *http.Request) (*http.Response, error) {
	t.lock.Lock()
	t.index = (t.index + 1) % len(t.Tokens)
	token := t.Tokens[t.index]
	t.lock.Unlock()

	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("Authorization", "Bearer "+token)
	return http.DefaultTransport.RoundTrip(reqClone)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
