/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package node_finder

import (
	"context"
	"fmt"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/clientcmd"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	"os"
)

type RdNodeFinder struct {
	KubeConfig             *restclient.Config
	KubeClient             *kubernetes.Clientset
	pod                    *core.Pod
	Namespace              string
	dbGoverningServiceName string
	RedisPort              int32
	redisTLSEnabled        bool
	dbClient               *cs.Clientset
	OffShootName           string
}

func New() *RdNodeFinder {
	var (
		rdTLSEnabled bool
	)
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		klog.Fatalln(err)
	}
	clientcmd.Fix(kubeConfig)
	dbClient, err := cs.NewForConfig(kubeConfig)
	if err != nil {
		klog.Fatalln(err)
	}
	namespace := os.Getenv("NAMESPACE")
	offShootName := os.Getenv("OFF_SHOOT_NAME")
	dbGoverningServiceName := os.Getenv("REDIS_GOVERNING_SERVICE")

	redisTLS := os.Getenv("REDIS_TLS")
	if redisTLS == "ON" {
		rdTLSEnabled = true
	}

	return &RdNodeFinder{
		KubeConfig:             kubeConfig,
		dbClient:               dbClient,
		Namespace:              namespace,
		OffShootName:           offShootName,
		RedisPort:              6379,
		redisTLSEnabled:        rdTLSEnabled,
		dbGoverningServiceName: dbGoverningServiceName,
	}
}

func (r *RdNodeFinder) RunRedisNodeFinder() {
	db, err := r.dbClient.KubedbV1alpha2().Redises(r.Namespace).Get(context.TODO(), r.OffShootName, metav1.GetOptions{})
	if err != nil {
		klog.Fatalln(err)
		return
	}
	dbMasterCount := int(*db.Spec.Cluster.Master)
	dbReplicaCount := int(*db.Spec.Cluster.Replicas)

	writeInfoToFile("master.txt", dbMasterCount)
	writeInfoToFile("replicas.txt", dbReplicaCount)

	var redisNodes []string
	for shardNo := 0; shardNo < dbMasterCount; shardNo++ {
		shardName := fmt.Sprintf("%s-shard%d", r.OffShootName, shardNo)

		for podNo := 0; podNo <= dbReplicaCount; podNo++ {
			podName := fmt.Sprintf("%s-%d", shardName, podNo)
			dnsName := podName + "." + r.dbGoverningServiceName
			redisNodes = append(redisNodes, dnsName)
		}
	}
	writePodDNSToFile("redis-nodes.txt", redisNodes)

	var masterNodes []string
	for shardNO := 0; shardNO < dbMasterCount; shardNO++ {
		initialMasterPod := fmt.Sprintf("%s-shard%d-0", r.OffShootName, shardNO)
		dnsName := initialMasterPod + "." + r.dbGoverningServiceName
		masterNodes = append(masterNodes, dnsName)
	}
	writePodDNSToFile("initial-master-nodes.txt", masterNodes)
}

func writeInfoToFile(filename string, count int) {

	filePath := fmt.Sprintf("/tmp/%s", filename)
	file, err := os.Create(filePath)
	if err != nil {
		klog.Fatalln(err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			klog.Fatalln(err)
		}
	}(file)
	_, err = file.WriteString(fmt.Sprintf("%d\n", count))
	if err != nil {
		klog.Fatalln(err)
	}

}

func writePodDNSToFile(filename string, dnsNames []string) {

	filePath := fmt.Sprintf("/tmp/%s", filename)
	file, err := os.Create(filePath)
	if err != nil {
		klog.Fatalln(err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			klog.Fatalln(err)
		}
	}(file)
	for _, dns := range dnsNames {
		_, err = file.WriteString(dns + "\n")
		if err != nil {
			klog.Fatalln(err)
			return
		}
	}
}
