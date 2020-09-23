package engineio

import (
	"github.com/googollee/go-socket.io/engineio/transport"
)

// HTTPError is error which has http response code
type HTTPError interface {
	Code() int
}

// Pauser is connection which can be paused and resumes.
type Pauser interface {
	Pause()
	Resume()
}

// Opener is client connection which need receive open message first.
type Opener interface {
	Open() (transport.ConnParams, error)
}

type TransportManager interface {
	Get(name string) transport.Transporter
	UpgradeFrom(name string) []string
}

// Manager is a manager of transports.
type Manager struct {
	order      []string

	transports map[string]transport.Transporter
}

// newTransportManager creates a new manager.
func newTransportManager(transports []transport.Transporter) *Manager {
	transpMap := make(map[string]transport.Transporter)
	names := make([]string, len(transports))

	for i, t := range transports {
		names[i] = t.Name()
		transpMap[t.Name()] = t
	}

	return &Manager{
		order:      names,
		transports: transpMap,
	}
}

// UpgradeFrom returns a name list of transports which can upgrade from given
// name.
func (m *Manager) UpgradeFrom(name string) []string {
	for i, n := range m.order {
		if n == name {
			return m.order[i+1:]
		}
	}
	return nil
}

// Get returns the transport with given name.
func (m *Manager) Get(name string) transport.Transporter {
	return m.transports[name]
}
