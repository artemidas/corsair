package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"ladon/backend/rule"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func evaluateStored(ctx context.Context, client kubernetes.Interface, rules []rule.Rule) []Finding {
	out := []Finding{}
	for _, r := range rules {
		findings, err := evaluateRule(ctx, client, r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error evaluating rule %s: %v\n", r.ID, err)
			continue
		}
		out = append(out, findings...)
	}
	return out
}

func evaluateRule(ctx context.Context, client kubernetes.Interface, r rule.Rule) ([]Finding, error) {
	items, err := fetchResourcesJSON(ctx, client, r.ResourceType)
	if err != nil {
		return nil, err
	}
	out := []Finding{}
	for _, item := range items {
		leaves := rule.EvaluateFieldPath(item, r.FieldPath)
		if !rule.EvaluateOperator(leaves, r.Operator, r.ExpectedValue) {
			continue
		}
		name, ns := metaFromJSON(item)
		msg := r.Title
		if r.Description != "" {
			msg = r.Title + ": " + r.Description
		}
		out = append(out, Finding{
			ID:           fmt.Sprintf("%s-%s-%s", r.ID, deref(ns), name),
			RuleID:       r.RuleID,
			RuleTitle:    r.Title,
			Severity:     r.Severity,
			ResourceKind: r.ResourceType,
			ResourceName: name,
			Namespace:    ns,
			Message:      msg,
		})
	}
	return out, nil
}

func fetchResourcesJSON(ctx context.Context, client kubernetes.Interface, resourceType string) ([]any, error) {
	switch resourceType {
	case "Pod":
		list, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return marshalItems(list.Items)
	case "ServiceAccount":
		list, err := client.CoreV1().ServiceAccounts("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return marshalItems(list.Items)
	case "Role":
		list, err := client.RbacV1().Roles("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return marshalItems(list.Items)
	case "ClusterRole":
		list, err := client.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return marshalItems(list.Items)
	case "RoleBinding":
		list, err := client.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return marshalItems(list.Items)
	case "ClusterRoleBinding":
		list, err := client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return marshalItems(list.Items)
	default:
		return []any{}, nil
	}
}

func marshalItems[T any](items []T) ([]any, error) {
	out := make([]any, 0, len(items))
	for i := range items {
		raw, err := json.Marshal(items[i])
		if err != nil {
			continue
		}
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			continue
		}
		out = append(out, v)
	}
	return out, nil
}

func metaFromJSON(item any) (name string, namespace *string) {
	obj, _ := item.(map[string]any)
	meta, _ := obj["metadata"].(map[string]any)
	if meta == nil {
		return "", nil
	}
	name, _ = meta["name"].(string)
	if ns, ok := meta["namespace"].(string); ok && ns != "" {
		return name, &ns
	}
	return name, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
