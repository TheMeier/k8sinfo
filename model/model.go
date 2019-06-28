package model

import (
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

// K8sInfoElement holds info data for one context
type K8sInfoElement struct {
	Deployments *apps.DeploymentList
	Services    *core.ServiceList
}

// K8sInfoData holds all scraped data for all contexts
type K8sInfoData map[string]*K8sInfoElement
