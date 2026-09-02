package cluster

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func writeKubeconfig(t *testing.T, current string, contexts ...string) string {
	t.Helper()
	cfg := api.NewConfig()
	cfg.CurrentContext = current
	cfg.Clusters["test"] = &api.Cluster{Server: "https://127.0.0.1:6443"}
	cfg.AuthInfos["test"] = &api.AuthInfo{}
	for _, name := range contexts {
		cfg.Contexts[name] = &api.Context{Cluster: "test", AuthInfo: "test"}
	}
	path := filepath.Join(t.TempDir(), "config")
	if err := clientcmd.WriteToFile(*cfg, path); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestListKubeContexts(t *testing.T) {
	path := writeKubeconfig(t, "minikube", "prod", "minikube")
	t.Setenv("KUBECONFIG", path)

	got, err := New(NewSession()).ListKubeContexts()
	if err != nil {
		t.Fatal(err)
	}
	if got.Current == nil || *got.Current != "minikube" {
		t.Fatalf("current = %v", got.Current)
	}
	if len(got.Contexts) != 2 || got.Contexts[0] != "minikube" || got.Contexts[1] != "prod" {
		t.Fatalf("contexts = %v", got.Contexts)
	}
}

func TestConnectProbeDisconnect(t *testing.T) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}}
	client := fake.NewSimpleClientset(ns)
	svc := newService(NewSession(), func(_ context.Context, name string) (kubernetes.Interface, error) {
		if name == "missing" {
			return nil, fmt.Errorf("context %q does not exist", name)
		}
		return client, nil
	}, defaultLoadKube)

	ctxName := "kind-dev"
	st, err := svc.ConnectCluster(&ctxName)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Connected || !st.Healthy || st.Context == nil || *st.Context != "kind-dev" {
		t.Fatalf("connect status = %+v", st)
	}

	probed, err := svc.ProbeCluster()
	if err != nil {
		t.Fatal(err)
	}
	if !probed.Connected || !probed.Healthy {
		t.Fatalf("probe = %+v", probed)
	}

	if _, _, err := svc.session.Client(); err != nil {
		t.Fatal(err)
	}

	gone := svc.DisconnectCluster()
	if gone.Connected || gone.Healthy || gone.Context != nil {
		t.Fatalf("disconnect = %+v", gone)
	}
	probed, err = svc.ProbeCluster()
	if err != nil {
		t.Fatal(err)
	}
	if probed.Connected {
		t.Fatalf("probe after disconnect = %+v", probed)
	}
}

func TestConnectFailsDoesNotStoreClient(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("list", "namespaces", func(ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("apiserver refused")
	})
	svc := newService(NewSession(), func(_ context.Context, _ string) (kubernetes.Interface, error) {
		return client, nil
	}, defaultLoadKube)

	name := "kind-dev"
	if _, err := svc.ConnectCluster(&name); err == nil {
		t.Fatal("expected list error")
	}
	if _, _, err := svc.session.Client(); err == nil {
		t.Fatal("client should not be stored after failed connect")
	}
}

func TestConnectUnknownContext(t *testing.T) {
	svc := newService(NewSession(), func(_ context.Context, name string) (kubernetes.Interface, error) {
		return nil, fmt.Errorf("context %q does not exist", name)
	}, defaultLoadKube)
	name := "missing"
	if _, err := svc.ConnectCluster(&name); err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveContextNameUsesKubeconfigCurrent(t *testing.T) {
	path := writeKubeconfig(t, "minikube", "minikube")
	t.Setenv("KUBECONFIG", path)
	svc := New(NewSession())
	if got := svc.resolveContextName(""); got != "minikube" {
		t.Fatalf("resolved = %q", got)
	}
	if got := svc.resolveContextName("  other  "); got != "other" {
		t.Fatalf("requested = %q", got)
	}
}
