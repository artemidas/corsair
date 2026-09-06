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
	prepared, err := rule.Prepare(r.Rego)
	if err != nil {
		return nil, fmt.Errorf("compile rule: %w", err)
	}
	out := []Finding{}
	for _, item := range items {
		msgs, err := prepared.Eval(ctx, item)
		if err != nil {
			return nil, err
		}
		if len(msgs) == 0 {
			continue
		}
		name, ns := metaFromJSON(item)
		for i, msg := range msgs {
			out = append(out, Finding{
				ID:           fmt.Sprintf("%s-%s-%s-%d", r.ID, deref(ns), name, i),
				RuleID:       r.RuleID,
				RuleTitle:    r.Title,
				Severity:     r.Severity,
				ResourceKind: r.ResourceType,
				ResourceName: name,
				Namespace:    ns,
				Message:      findingMessage(r, msg),
			})
		}
	}
	return out, nil
}

func findingMessage(r rule.Rule, msg string) string {
	if msg != "" {
		return msg
	}
	if r.Description != "" {
		return r.Title + ": " + r.Description
	}
	return r.Title
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
