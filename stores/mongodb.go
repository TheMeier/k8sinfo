package stores

import (
	"fmt"

	"github.com/TheMeier/k8sinfo/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	log "github.com/sirupsen/logrus"
)

// UpdateMongoDB writes all data to mongodb
func UpdateMongoDB(data model.K8sInfoData, session *mgo.Session) {
	db := session.DB("k8sinfo")
	now := bson.Now()
	for context, element := range data {
		for _, deployment := range element.Deployments.Items {

			name := fmt.Sprintf("%s_%s_%s", context, deployment.Namespace, deployment.Name)
			filter := bson.M{"name": name}
			update := model.DeploymentElement{
				Name:       name,
				Deployment: &deployment,
				Timestamp:  now,
			}
			info, err := db.C("deployments").Upsert(filter, &update)
			if err != nil {
				log.Errorf("Failed to insert %+v %+v", err, info)
			}

		}

		for _, service := range element.Services.Items {

			name := fmt.Sprintf("%s_%s_%s", context, service.Namespace, service.Name)
			filter := bson.M{"name": name}
			update := model.ServiceElement{
				Name:      name,
				Service:   &service,
				Timestamp: now,
			}
			info, err := db.C("services").Upsert(filter, &update)
			if err != nil {
				log.Errorf("Failed to insert %+v %+v", err, info)
			}
		}

		for _, ingress := range element.Ingresses.Items {

			name := fmt.Sprintf("%s_%s_%s", context, ingress.Namespace, ingress.Name)
			filter := bson.M{"name": name}
			update := model.IngressElement{
				Name:      name,
				Ingress:   &ingress,
				Timestamp: now,
			}
			info, err := db.C("ingresses").Upsert(filter, &update)
			if err != nil {
				log.Errorf("Failed to insert %+v %+v", err, info)
			}
		}
	}
}
