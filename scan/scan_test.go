package scan

import (
	"fmt"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	"ladon/appdb"
	"ladon/cluster"
	"ladon/project"
	"ladon/rule"
)

func boolPtr(v bool) *bool { return &v }

func privilegedPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "evil", Namespace: "default"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "main",
				SecurityContext: &corev1.SecurityContext{
					Privileged: boolPtr(true),
				},
			}},
		},
	}
}

func setup(t *testing.T) (*Service, *cluster.Session, *project.Service) {
	t.Helper()
	db, err := appdb.Open(filepath.Join(t.TempDir(), "ladon.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = appdb.Close(db) })
	session := cluster.NewSession()
	return New(db, session, rule.New(db)), session, project.New(db)
}

func TestPreviewAndRunScan(t *testing.T) {
	svc, session, projects := setup(t)

	if _, err := svc.PreviewScan(); err == nil {
		t.Fatal("preview should fail when disconnected")
	}

	session.Set(fake.NewSimpleClientset(privilegedPod()), "kind-dev")

	preview, err := svc.PreviewScan()
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(preview, "POD001") {
		t.Fatalf("expected POD001 in preview, got %v", ruleIDs(preview))
	}
	if !hasTitle(preview, "Privileged container") {
		t.Fatalf("expected stored privileged finding, got %v", titles(preview))
	}

	created, err := projects.CreateProject(project.ProjectInput{
		Name: "demo",
		Kind: project.KindKubernetesClusterReview,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.RunScan("missing"); err == nil {
		t.Fatal("expected missing project")
	}

	result, err := svc.RunScan(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Scan.Status != StatusCompleted || result.Scan.FindingCount == 0 {
		t.Fatalf("run = %+v", result.Scan)
	}
	if result.Scan.Context == nil || *result.Scan.Context != "kind-dev" {
		t.Fatalf("context = %v", result.Scan.Context)
	}

	listed, err := svc.ListScans(created.ID)
	if err != nil || len(listed) != 1 || listed[0].ID != result.Scan.ID {
		t.Fatalf("list = %+v %v", listed, err)
	}
	got, err := svc.GetScan(result.Scan.ID)
	if err != nil || got == nil || got.ID != result.Scan.ID {
		t.Fatalf("get = %+v %v", got, err)
	}
	findings, err := svc.ListScanFindings(result.Scan.ID)
	if err != nil || int64(len(findings)) != result.Scan.FindingCount {
		t.Fatalf("findings = %d %v", len(findings), err)
	}

	if err := projects.DeleteProject(created.ID); err != nil {
		t.Fatal(err)
	}
	listed, err = svc.ListScans(created.ID)
	if err != nil || len(listed) != 0 {
		t.Fatalf("scans after delete = %+v %v", listed, err)
	}
}

func TestRunScanPersistsFailure(t *testing.T) {
	svc, session, projects := setup(t)
	created, err := projects.CreateProject(project.ProjectInput{
		Name: "demo",
		Kind: project.KindKubernetesClusterReview,
	})
	if err != nil {
		t.Fatal(err)
	}

	client := fake.NewSimpleClientset()
	client.PrependReactor("list", "pods", func(ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("apiserver down")
	})
	session.Set(client, "kind-dev")

	result, err := svc.RunScan(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Scan.Status != StatusFailed || result.Scan.Error == nil {
		t.Fatalf("failed scan = %+v", result.Scan)
	}
}

func TestHardcodedRBAC001(t *testing.T) {
	crb := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "bind-admin"},
		RoleRef:    rbacv1.RoleRef{Name: "cluster-admin"},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      "default",
			Namespace: "kube-system",
		}},
	}
	got := rbac001([]rbacv1.ClusterRoleBinding{crb})
	if len(got) != 1 || got[0].RuleID != "RBAC001" || deref(got[0].Namespace) != "kube-system" {
		t.Fatalf("rbac001 = %+v", got)
	}
}

func hasRule(findings []Finding, id string) bool {
	for _, f := range findings {
		if f.RuleID == id {
			return true
		}
	}
	return false
}

func hasTitle(findings []Finding, title string) bool {
	for _, f := range findings {
		if f.RuleTitle == title {
			return true
		}
	}
	return false
}

func ruleIDs(findings []Finding) []string {
	out := make([]string, len(findings))
	for i, f := range findings {
		out[i] = f.RuleID
	}
	return out
}

func titles(findings []Finding) []string {
	out := make([]string, len(findings))
	for i, f := range findings {
		out[i] = f.RuleTitle
	}
	return out
}
