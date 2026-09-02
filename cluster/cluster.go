package cluster

import (
	"strings"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type ClusterStatus struct {
	Connected bool    `json:"connected"`
	Healthy   bool    `json:"healthy"`
	Context   *string `json:"context"`
	Error     *string `json:"error"`
}

type KubeContexts struct {
	Current  *string  `json:"current"`
	Contexts []string `json:"contexts"`
}

// Session holds the in-memory cluster client for this process.
// Other Go services (scan) can share it; it is not a Wails service.
type Session struct {
	mu      sync.Mutex
	client  kubernetes.Interface
	context string
}

func NewSession() *Session {
	return &Session{}
}

func (s *Session) set(client kubernetes.Interface, contextName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client = client
	s.context = contextName
}

func (s *Session) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client = nil
	s.context = ""
}

func (s *Session) snapshot() (kubernetes.Interface, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.client, s.context
}

func (s *Session) Client() (kubernetes.Interface, string, error) {
	client, contextName := s.snapshot()
	if client == nil {
		return nil, "", errNotConnected
	}
	return client, contextName, nil
}

func disconnected() ClusterStatus {
	return ClusterStatus{}
}

func connectedStatus(contextName string) ClusterStatus {
	return ClusterStatus{
		Connected: true,
		Healthy:   true,
		Context:   strPtr(contextName),
	}
}

func unreachable(contextName string, err error) ClusterStatus {
	msg := err.Error()
	return ClusterStatus{
		Connected: true,
		Healthy:   false,
		Context:   strPtr(contextName),
		Error:     &msg,
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func defaultLoadKube() (*api.Config, error) {
	return clientcmd.NewDefaultClientConfigLoadingRules().Load()
}
