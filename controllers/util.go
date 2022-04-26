package controllers

import (
	networkingv1 "k8s.io/api/networking/v1"
)

func extractHosts(i *networkingv1.Ingress) []string {
	ret := []string{}
	for _, rule := range i.Spec.Rules {
		if rule.Host != "" {
			ret = append(ret, rule.Host)
		}
	}
	return ret
}
