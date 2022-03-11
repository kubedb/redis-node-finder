package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/clientcmd"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
)

var (
	//err        error
	//hostName   string
	kubeConfig *restclient.Config
	//kubeClient *kubernetes.Clientset
	dbClient *cs.Clientset
)

func init() {
	var err error
	kubeConfig, err = restclient.InClusterConfig()
	if err != nil {
		klog.Fatalln(err)
	}
	clientcmd.Fix(kubeConfig)
	dbClient, err = cs.NewForConfig(kubeConfig)
	if err != nil {
		klog.Fatalln(err)
	}
}

func main() {
	fmt.Println("Hello")
	findRedisNodes()
}

func findRedisNodes() {
	nameSpace := "demo"
	db, err := dbClient.KubedbV1alpha2().Redises(nameSpace).Get(context.TODO(), "dbname", metav1.GetOptions{})
	if err != nil {
		klog.Fatalln(err)
		return
	}
	dbReplicaCount := int(*db.Spec.Cluster.Replicas)
	fmt.Println(dbReplicaCount)
}
