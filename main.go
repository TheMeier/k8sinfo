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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func scrapeData(kubeconfigs []string) {

	var newData model.K8sInfoData
	for _, kubeconfig := range kubeconfigs {
		cnf, err := clientcmd.LoadFromFile(kubeconfig)
		if err != nil {
			log.Errorf("Failed to parse config at %s", kubeconfig)
			log.Errorf("%s", err)
			return
		}

		for contextName := range cnf.Contexts {
			log.Debugf("Context: %s", contextName)

			override := getDefaultOverride()
			config := clientcmd.NewNonInteractiveClientConfig(*cnf, contextName, &override, nil)
			clientConfig, err := config.ClientConfig()
			if err != nil {
				log.Errorf("Failed to create clientConfig: %s", err)
				continue
			}
			clientset, err := kubernetes.NewForConfig(clientConfig)
			deployments, err := clientset.Apps().Deployments("").List(v1.ListOptions{})
			if err != nil {
				log.Errorf("Failed to list deployments: %s", err)
				continue
			}
			for _, deployment := range deployments.Items {
				newData.Deployments = append(newData.Deployments, deployment)
			}
			services, err := clientset.CoreV1().Services("").List(v1.ListOptions{})
			if err != nil {
				log.Errorf("Failed to list services: %s", err)
				continue
			}
			for _, service := range services.Items {
				newData.Services = append(newData.Services, service)
			}

		}
	}
	k8sInfoData.Set(newData)
}

func k8sHTTPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(k8sInfoData.Get())
}
func k8sHTTPHandlerDeployments(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(k8sInfoData.Get().Deployments)
}
func k8sHTTPHandlerServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(k8sInfoData.Get().Services)
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	kubeconfigs := kingpin.Flag("kubeconfig",
		"path to one or multiple kubeconfig files").
		Default(filepath.Join(homedir.HomeDir(), ".kube", "config")).
		Short('c').
		ExistingFiles()
	scrapeInterval := kingpin.Flag("scrapeInterval",
		"Interval between data scraping").
		Default("120").
		Short('i').
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
	log.Infof("Staring k8sinfo, listening on %s, scrape interval %d",
		*host,
		*scrapeInterval)

	scrapeData(*kubeconfigs)
	go func() {
		gocron.Every(uint64(*scrapeInterval)).Seconds().Do(scrapeData, *kubeconfigs)
		<-gocron.Start()
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", k8sHTTPHandler)
	http.HandleFunc("/deployments", k8sHTTPHandlerDeployments)
	http.HandleFunc("/services", k8sHTTPHandlerServices)
	log.Fatal(http.ListenAndServe(*host, nil))

}
