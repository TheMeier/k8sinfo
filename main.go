package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/TheMeier/k8sinfo/model"
	"github.com/jasonlvhit/gocron"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

var k8sInfoData = NewK8sInfoHolder()

func getDefaultOverride() clientcmd.ConfigOverrides {
	return clientcmd.ConfigOverrides{
		ClusterInfo: api.Cluster{
			Server: "",
		},
	}
}
func scrapeData(kubeconfig string) {

	cnf, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		log.Errorf("Failed to parse config at %s", kubeconfig)
		log.Errorf("%s", err)
		return
	}
	var newData model.K8sInfoData

	for contextName := range cnf.Contexts {
		log.Debugf("Context: %s\n", contextName)

		override := getDefaultOverride()
		config := clientcmd.NewNonInteractiveClientConfig(*cnf, contextName, &override, nil)
		clientConfig, err := config.ClientConfig()
		if err != nil {
			log.Errorf("Failed to create clientConfig: %s", err)
			continue
		}
		clientset, err := kubernetes.NewForConfig(clientConfig)
		deployments, err := clientset.AppsV1beta1().Deployments("").List(v1.ListOptions{})
		if err != nil {
			log.Errorf("Failed to list deployments: %s", err)
			continue
		}
		for _, deployment := range deployments.Items {

			for _, initContainter := range deployment.Spec.Template.Spec.InitContainers {
				log.Debugf("Initcontainer: %s, %s", initContainter.Name, deployment.Namespace)
				newData.Deployments = append(newData.Deployments,
					model.DeploymentData{Namespace: deployment.Namespace, Image: initContainter.Image, Name: initContainter.Name, Context: contextName})
			}
			for _, containter := range deployment.Spec.Template.Spec.Containers {
				log.Debugf("Container: %s, %s", containter.Name, deployment.Namespace)
				newData.Deployments = append(newData.Deployments,
					model.DeploymentData{Namespace: deployment.Namespace, Image: containter.Image, Name: containter.Name, Context: contextName})

			}

			log.Debugf("%#v\n", newData)
		}

		k8sInfoData.Set(newData)
	}
}

func k8sHTTPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(k8sInfoData.Get())
}

func main() {
	var kubeconfig *string
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	kubeconfig = kingpin.Flag("kubeconfig",
		"absolute path to the kubeconfig file").
		Default(filepath.Join(homedir.HomeDir(), ".kube", "config")).
		Short('c').
		String()
	scrapeInterval := kingpin.Flag("scarpeInterval",
		"Interval for between data scarping").
		Default("2").
		Int()
	host := kingpin.Flag("web.listen-address",
		"Address to listen on for http requests").
		Default(":2112").
		Short('l').
		String()
	debug := kingpin.Flag("debug", "Set log level to debug").
		Default("false").
		Short('d').
		Bool()
	kingpin.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	go func() {
		gocron.Every(uint64(*scrapeInterval)).Seconds().Do(scrapeData, *kubeconfig)
		<-gocron.Start()
	}()

	log.Infof("Staring k8sinfo, listening on %s, scrape interval %d",
		*host,
		*scrapeInterval)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", k8sHTTPHandler)
	log.Fatal(http.ListenAndServe(*host, nil))

}