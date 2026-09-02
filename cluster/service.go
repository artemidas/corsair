package cluster

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var errNotConnected = errors.New("not connected to a cluster")

type dialFunc func(ctx context.Context, contextName string) (kubernetes.Interface, error)
type kubeLoader func() (*api.Config, error)

type Service struct {
	session  *Session
	dial     dialFunc
	loadKube kubeLoader
}

func New(session *Session) *Service {
	return newService(session, defaultDial, defaultLoadKube)
}

func newService(session *Session, dial dialFunc, load kubeLoader) *Service {
	if session == nil {
		session = NewSession()
	}
	if dial == nil {
		dial = defaultDial
	}
	if load == nil {
		load = defaultLoadKube
	}
	return &Service{session: session, dial: dial, loadKube: load}
}

func (s *Service) ListKubeContexts() (KubeContexts, error) {
	cfg, err := s.loadKube()
	if err != nil {
		return KubeContexts{}, fmt.Errorf("read kubeconfig: %w", err)
	}
	names := make([]string, 0, len(cfg.Contexts))
	for name := range cfg.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)
	return KubeContexts{
		Current:  strPtr(cfg.CurrentContext),
		Contexts: names,
	}, nil
}

func (s *Service) ConnectCluster(kubeContext *string) (ClusterStatus, error) {
	requested := deref(kubeContext)
	client, err := s.dial(context.Background(), requested)
	if err != nil {
		return ClusterStatus{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{}); err != nil {
		return ClusterStatus{}, fmt.Errorf("list namespaces: %w", err)
	}
	resolved := s.resolveContextName(requested)
	s.session.Set(client, resolved)
	return connectedStatus(resolved), nil
}

func (s *Service) DisconnectCluster() ClusterStatus {
	s.session.clear()
	return disconnected()
}

func (s *Service) ProbeCluster() (ClusterStatus, error) {
	client, contextName := s.session.snapshot()
	if client == nil {
		return disconnected(), nil
	}
	if _, err := client.Discovery().ServerVersion(); err != nil {
		return unreachable(contextName, err), nil
	}
	return connectedStatus(contextName), nil
}

func (s *Service) resolveContextName(requested string) string {
	if name := strings.TrimSpace(requested); name != "" {
		return name
	}
	cfg, err := s.loadKube()
	if err != nil {
		return ""
	}
	return cfg.CurrentContext
}

func defaultDial(_ context.Context, contextName string) (kubernetes.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	if name := strings.TrimSpace(contextName); name != "" {
		overrides.CurrentContext = name
	}
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	cfg.Timeout = 15 * time.Second
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client: %w", err)
	}
	return client, nil
}
