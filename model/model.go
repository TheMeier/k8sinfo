package model

import (
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

// K8sInfoData holds all scraped data
type K8sInfoData struct {
	Deployments []apps.Deployment
	Services    []core.Service
}
