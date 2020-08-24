package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/TheMeier/k8sinfo/model"
	"github.com/TheMeier/k8sinfo/stores"
	"github.com/globalsign/mgo"
	"github.com/jasonlvhit/gocron"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
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
func scrapeData(kubeconfigs []string, mongoSession *mgo.Session, mongoEnable *bool) {

	newData := make(map[string]*model.K8sInfoElement)
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
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			clientset, err := kubernetes.NewForConfig(clientConfig)
			deployments, _ := clientset.AppsV1().Deployments("").List(ctx, v1.ListOptions{})
			services, _ := clientset.CoreV1().Services("").List(ctx, v1.ListOptions{})
			ingresses, _ := clientset.ExtensionsV1beta1().Ingresses("").List(ctx, v1.ListOptions{})
			newData[contextName] = &model.K8sInfoElement{
				Deployments: deployments,
				Services:    services,
				Ingresses:   ingresses,
			}
		}
	}
	k8sInfoData.Set(newData)
	if *mongoEnable {
		stores.UpdateMongoDB(k8sInfoData.Get(), mongoSession)
	}
}

func k8sHTTPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(k8sInfoData.Get())
}

func k8sHTTPHandlerDeployments(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	data := k8sInfoData.Get()
	ret := make(map[string]*apps.DeploymentList)
	for context, value := range data {
		ret[context] = value.Deployments
	}
	json.NewEncoder(w).Encode(ret)
}

func k8sHTTPHandlerServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	data := k8sInfoData.Get()
	ret := make(map[string]*core.ServiceList)
	for context, value := range data {
		ret[context] = value.Services
	}
	json.NewEncoder(w).Encode(ret)
}

func k8sHTTPHandlerIngresses(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	data := k8sInfoData.Get()
	ret := make(map[string]*v1beta1.IngressList)
	for context, value := range data {
		ret[context] = value.Ingresses
	}
	json.NewEncoder(w).Encode(ret)
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
	mongoEnable := kingpin.Flag("mongoEnable", "Enable exporter for mongodb").
		Default("false").
		Bool()
	mongoAddress := kingpin.Flag("mongoAddress",
		"address to mongo seed servers, can be specified multiple times").
		Default("localhost:27017").
		Short('m').
		Strings()

	kingpin.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	log.Infof("Staring k8sinfo, listening on %s, scrape interval %d",
		*host,
		*scrapeInterval)

	mongoSession := &mgo.Session{}

	if *mongoEnable {
		mongoDBDialInfo := &mgo.DialInfo{
			Addrs:    *mongoAddress,
			Timeout:  60 * time.Second,
			Database: "k8sinfo",
			Username: os.Getenv("MONGO_USERNAME"),
			Password: os.Getenv("MONGO_PASSWORD"),
		}
		var err error
		mongoSession, err = mgo.DialWithInfo(mongoDBDialInfo)
		if err != nil {
			log.Fatalf("CreateSession: %s\n", err)
		}
	} else {
		mongoSession = &mgo.Session{}
	}

	scrapeData(*kubeconfigs, mongoSession, mongoEnable)
	go func() {
		gocron.Every(uint64(*scrapeInterval)).Seconds().Do(scrapeData, *kubeconfigs, mongoSession, mongoEnable)
		<-gocron.Start()
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", k8sHTTPHandler)
	http.HandleFunc("/deployments", k8sHTTPHandlerDeployments)
	http.HandleFunc("/services", k8sHTTPHandlerServices)
	http.HandleFunc("/ingresses", k8sHTTPHandlerIngresses)
	http.HandleFunc("/trigger", func(w http.ResponseWriter, r *http.Request) {
		scrapeData(*kubeconfigs, mongoSession, mongoEnable)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
	})
	log.Fatal(http.ListenAndServe(*host, nil))

}
