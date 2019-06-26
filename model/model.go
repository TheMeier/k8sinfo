package model

import (
	v1beta1 "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
)

// K8sInfoData holds all scraped data
type K8sInfoData struct {
	Deployments []v1beta1.Deployment
	Services    []core.Service
}
