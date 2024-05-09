package transmission

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

var _ http.RoundTripper = &authenticator{}

type authenticator struct {
	next      http.RoundTripper
	lock      sync.RWMutex
	sessionID string
}

const transmissionSessionIDHeader = "X-Transmission-Session-Id"

func (a *authenticator) RoundTrip(request *http.Request) (*http.Response, error) {
	var bodyCopy bytes.Buffer
	origBody := io.TeeReader(request.Body, &bodyCopy)
	request.Body = io.NopCloser(origBody)

	request.Header.Set(transmissionSessionIDHeader, a.getSessionID())

	resp, err := a.next.RoundTrip(request)
	if err != nil || resp.StatusCode != http.StatusConflict {
		return resp, err
	}

	if resp.Body != nil {
		_ = resp.Body.Close()
	}

	sessionID := resp.Header.Get(transmissionSessionIDHeader)
	a.setSessionID(sessionID)
	request.Header.Set(transmissionSessionIDHeader, sessionID)
	request.Body = io.NopCloser(&bodyCopy)

	return a.next.RoundTrip(request)
}

func (a *authenticator) getSessionID() string {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.sessionID
}

func (a *authenticator) setSessionID(sessionID string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.sessionID = sessionID
}
