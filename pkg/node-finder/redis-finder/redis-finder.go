/*
Copyright AppsCode Inc. and Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package redis_finder

import (
	"context"
	"fmt"
	"os"

	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/clientcmd"
)

type RedisdNodeFinder struct {
	Namespace              string
	dbGoverningServiceName string
	RedisPort              int32
	dbClient               *cs.Clientset
	RedisName              string
	masterFile             string
	slaveFile              string
	redisNodesFile         string
	initialMasterNodesFile string
}

func New(masterFile string, slaveFile string, redisNodesFile string, initialMasterNodesFile string) *RedisdNodeFinder {
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
	RedisName := os.Getenv("REDIS_NAME")
	dbGoverningServiceName := os.Getenv("REDIS_GOVERNING_SERVICE")

	return &RedisdNodeFinder{
		dbClient:               dbClient,
		Namespace:              namespace,
		RedisName:              RedisName,
		RedisPort:              6379,
		dbGoverningServiceName: dbGoverningServiceName,
		masterFile:             masterFile,
		slaveFile:              slaveFile,
		redisNodesFile:         redisNodesFile,
		initialMasterNodesFile: initialMasterNodesFile,
	}
}

// RunRedisNodeFinder  get Redis DB  object and extract shard/replica count, and initial master nodes name and write them to given
// file name in /tmp directory. The call is made from init script, so it will write to tmp/ directory
// The init script then use those value to provision the db object with right configuration and the init
// script also has updated information during pod restart
func (r *RedisdNodeFinder) RunRedisNodeFinder() {
	db, err := r.dbClient.KubedbV1().Redises(r.Namespace).Get(context.TODO(), r.RedisName, metav1.GetOptions{})
	if err != nil {
		klog.Fatalln(err)
		return
	}
	dbShardCount := int(*db.Spec.Cluster.Shards)
	dbReplicaCount := int(*db.Spec.Cluster.Replicas)

	r.writeInfoToFile(r.masterFile, dbShardCount)
	r.writeInfoToFile(r.slaveFile, dbReplicaCount-1)

	var redisNodes []string
	for shardNo := 0; shardNo < dbShardCount; shardNo++ {
		shardName := fmt.Sprintf("%s-shard%d", r.RedisName, shardNo)

		for podNo := 0; podNo < dbReplicaCount; podNo++ {
			podName := fmt.Sprintf("%s-%d", shardName, podNo)
			dnsName := podName + "." + r.dbGoverningServiceName
			redisNodes = append(redisNodes, dnsName)
		}
	}
	r.writePodDNSToFile(r.redisNodesFile, redisNodes)

	var masterNodes []string
	for shardNO := 0; shardNO < dbShardCount; shardNO++ {
		initialMasterPod := fmt.Sprintf("%s-shard%d-0", r.RedisName, shardNO)
		dnsName := initialMasterPod + "." + r.dbGoverningServiceName
		masterNodes = append(masterNodes, dnsName)
	}
	r.writePodDNSToFile(r.initialMasterNodesFile, masterNodes)
}

func (r *RedisdNodeFinder) writeInfoToFile(filename string, count int) {
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

func (r *RedisdNodeFinder) writePodDNSToFile(filename string, dnsNames []string) {
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
