package model

import (
	"github.com/globalsign/mgo/bson"
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

//DeploymentElement wraps a deployment for storage in a backend
type DeploymentElement struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	Deployment *apps.Deployment
	Name       string `json:"name"`
}

//ServiceElement wraps a deployment for storage in a backend
type ServiceElement struct {
	ID      bson.ObjectId `bson:"_id,omitempty"`
	Service *core.Service
	Name    string `json:"name"`
}
