package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
)

// K8sInfoElement holds info data for one context
type K8sInfoElement struct {
	Deployments *apps.DeploymentList
	Services    *core.ServiceList
	Ingresses   *v1beta1.IngressList
}

// K8sInfoData holds all scraped data for all contexts
type K8sInfoData map[string]*K8sInfoElement

//DeploymentElement wraps a deployment for storage in a backend
type DeploymentElement struct {
	ID         bson.ObjectId    `bson:"_id,omitempty"`
	Deployment *apps.Deployment `bson:",inline"`
	Name       string           `bson:"name"`
	Timestamp  time.Time        `bson:"timestamp"`
}

//ServiceElement wraps a deployment for storage in a backend
type ServiceElement struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Service   *core.Service `bson:",inline"`
	Name      string        `bson:"name"`
	Timestamp time.Time     `bson:"timestamp"`
}

//IngressElement wraps a deployment for storage in a backend
type IngressElement struct {
	ID        bson.ObjectId    `bson:"_id,omitempty"`
	Ingress   *v1beta1.Ingress `bson:",inline"`
	Name      string           `bson:"name"`
	Timestamp time.Time        `bson:"timestamp"`
}
