package scan

import (
	"context"
	"fmt"

	"ladon/rule"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func runHardcoded(ctx context.Context, client kubernetes.Interface) ([]Finding, error) {
	crbs, err := client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	roles, err := client.RbacV1().Roles("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	clusterRoles, err := client.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := []Finding{}
	out = append(out, rbac001(crbs.Items)...)
	out = append(out, rbac002Roles(roles.Items)...)
	out = append(out, rbac002ClusterRoles(clusterRoles.Items)...)
	out = append(out, pod001(pods.Items)...)
	out = append(out, pod004(pods.Items)...)
	return out, nil
}

func rbac001(crbs []rbacv1.ClusterRoleBinding) []Finding {
	out := []Finding{}
	for _, crb := range crbs {
		if crb.RoleRef.Name != "cluster-admin" {
			continue
		}
		for _, subject := range crb.Subjects {
			if subject.Kind != "ServiceAccount" || subject.Name != "default" {
				continue
			}
			ns := subject.Namespace
			if ns == "" {
				ns = "default"
			}
			crbName := crb.Name
			out = append(out, Finding{
				ID:           fmt.Sprintf("RBAC001-%s-%s", ns, crbName),
				RuleID:       "RBAC001",
				RuleTitle:    "Default ServiceAccount bound to cluster-admin",
				Severity:     rule.SeverityCritical,
				ResourceKind: "ServiceAccount",
				ResourceName: "default",
				Namespace:    &ns,
				Message: fmt.Sprintf(
					"ServiceAccount 'default' in namespace '%s' is bound to cluster-admin via ClusterRoleBinding '%s'",
					ns, crbName,
				),
			})
		}
	}
	return out
}

func rbac002Roles(roles []rbacv1.Role) []Finding {
	out := []Finding{}
	for _, r := range roles {
		if !hasWildcardVerb(r.Rules) {
			continue
		}
		name := r.Name
		ns := r.Namespace
		var nsPtr *string
		if ns != "" {
			nsPtr = &ns
		}
		out = append(out, Finding{
			ID:           fmt.Sprintf("RBAC002-Role-%s-%s", ns, name),
			RuleID:       "RBAC002",
			RuleTitle:    "Role grants wildcard verb",
			Severity:     rule.SeverityHigh,
			ResourceKind: "Role",
			ResourceName: name,
			Namespace:    nsPtr,
			Message:      fmt.Sprintf("Role '%s' grants wildcard verb '*'", name),
		})
	}
	return out
}

func rbac002ClusterRoles(roles []rbacv1.ClusterRole) []Finding {
	out := []Finding{}
	for _, cr := range roles {
		if !hasWildcardVerb(cr.Rules) {
			continue
		}
		name := cr.Name
		out = append(out, Finding{
			ID:           fmt.Sprintf("RBAC002-ClusterRole-%s", name),
			RuleID:       "RBAC002",
			RuleTitle:    "Role grants wildcard verb",
			Severity:     rule.SeverityHigh,
			ResourceKind: "ClusterRole",
			ResourceName: name,
			Message:      fmt.Sprintf("ClusterRole '%s' grants wildcard verb '*'", name),
		})
	}
	return out
}

func pod001(pods []corev1.Pod) []Finding {
	out := []Finding{}
	for _, pod := range pods {
		name := pod.Name
		ns := pod.Namespace
		var nsPtr *string
		if ns != "" {
			nsPtr = &ns
		}
		for _, c := range pod.Spec.Containers {
			if c.SecurityContext == nil || c.SecurityContext.Privileged == nil || !*c.SecurityContext.Privileged {
				continue
			}
			out = append(out, Finding{
				ID:           fmt.Sprintf("POD001-%s-%s-%s", ns, name, c.Name),
				RuleID:       "POD001",
				RuleTitle:    "Privileged container",
				Severity:     rule.SeverityCritical,
				ResourceKind: "Pod",
				ResourceName: name,
				Namespace:    nsPtr,
				Message: fmt.Sprintf(
					"Container '%s' in Pod '%s' runs with securityContext.privileged=true",
					c.Name, name,
				),
			})
		}
	}
	return out
}

func pod004(pods []corev1.Pod) []Finding {
	out := []Finding{}
	for _, pod := range pods {
		name := pod.Name
		ns := pod.Namespace
		var nsPtr *string
		if ns != "" {
			nsPtr = &ns
		}
		var podLevel *bool
		if pod.Spec.SecurityContext != nil {
			podLevel = pod.Spec.SecurityContext.RunAsNonRoot
		}
		for _, c := range pod.Spec.Containers {
			effective := podLevel
			if c.SecurityContext != nil && c.SecurityContext.RunAsNonRoot != nil {
				effective = c.SecurityContext.RunAsNonRoot
			}
			if effective != nil && *effective {
				continue
			}
			out = append(out, Finding{
				ID:           fmt.Sprintf("POD004-%s-%s-%s", ns, name, c.Name),
				RuleID:       "POD004",
				RuleTitle:    "Container not running as non-root",
				Severity:     rule.SeverityMedium,
				ResourceKind: "Pod",
				ResourceName: name,
				Namespace:    nsPtr,
				Message: fmt.Sprintf(
					"Container '%s' in Pod '%s' does not set securityContext.runAsNonRoot=true",
					c.Name, name,
				),
			})
		}
	}
	return out
}

func hasWildcardVerb(rules []rbacv1.PolicyRule) bool {
	for _, r := range rules {
		for _, v := range r.Verbs {
			if v == "*" {
				return true
			}
		}
	}
	return false
}
