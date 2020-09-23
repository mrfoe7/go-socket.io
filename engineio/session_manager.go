package engineio

import (
	"strconv"
	"sync"
	"sync/atomic"
)

// SessionIDGenerator generates new session id. Default behavior is simple increasing number.
// If you need custom session id, for example using local ip as perfix, you can
// implement SessionIDGenerator and save in Configure. Engine.io will use custom
// one to generate new session id.
type SessionIDGenerator interface {
	NewID() string
}

type defaultIDGenerator struct {
	nextID uint64
}

func newDefaultIDGenerator() *defaultIDGenerator{
	return &defaultIDGenerator{}
}

func (g *defaultIDGenerator) NewID() string {
	id := atomic.AddUint64(&g.nextID, 1)
	return strconv.FormatUint(id, 36)
}

// SessionManager
type SessionManager interface {
	SessionIDGenerator

	Add(Session)
	Get(string) Session
	Remove(string)
	Count() int
}

type manager struct {
	SessionIDGenerator

	sessions      map[string]Session

	lock sync.RWMutex
}

func newSessionManager(gen SessionIDGenerator) *manager {
	if gen == nil {
		gen = newDefaultIDGenerator()
	}

	return &manager{
		SessionIDGenerator: gen,
		sessions:         	make(map[string]Session),
	}
}

func (m *manager) Add(s Session) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.sessions[s.GetID()] = s
}

func (m *manager) Get(sid string) Session {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.sessions[sid]
}

func (m *manager) Remove(sid string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.sessions[sid]; !ok {
		return
	}

	delete(m.sessions, sid)
}

func (m *manager) Count() int {
	m.lock.Lock()
	defer m.lock.Unlock()

	return len(m.sessions)
}
