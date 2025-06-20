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
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	v1 "kubedb.dev/apimachinery/apis/kubedb/v1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v2 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/clientcmd"
	psc "kubeops.dev/petset/client/clientset/versioned"
)

type RedisdNodeFinder struct {
	Namespace              string
	dbGoverningServiceName string
	RedisPort              int32
	dbClient               *cs.Clientset
	psClient               *psc.Clientset
	coreV1Client           *v2.CoreV1Client
	RedisName              string
	masterFile             string
	slaveFile              string
	NodesFile              string
	initialMasterNodesFile string
	endpointTypeFile       string
}

func New(masterFile string, slaveFile string, nodesFile string, initialMasterNodesFile string, endpointTypeFile string) *RedisdNodeFinder {
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		klog.Fatalln(err)
	}
	clientcmd.Fix(kubeConfig)
	dbClient, err := cs.NewForConfig(kubeConfig)
	if err != nil {
		klog.Fatalln(err)
	}

	psClient, err := psc.NewForConfig(kubeConfig)
	if err != nil {
		klog.Fatalln(err)
	}

	coreV1Client, err := v2.NewForConfig(kubeConfig)
	if err != nil {
		klog.Fatalln(err)
	}
	time.Sleep(time.Second * 5)

	namespace := os.Getenv("NAMESPACE")

	envKeyDbName := "DATABASE_NAME"
	envKeyGovService := "DATABASE_GOVERNING_SERVICE"
	RedisName := os.Getenv(envKeyDbName)
	dbGoverningServiceName := os.Getenv(envKeyGovService)

	return &RedisdNodeFinder{
		dbClient:               dbClient,
		psClient:               psClient,
		coreV1Client:           coreV1Client,
		Namespace:              namespace,
		RedisName:              RedisName,
		dbGoverningServiceName: dbGoverningServiceName,
		masterFile:             masterFile,
		slaveFile:              slaveFile,
		NodesFile:              nodesFile,
		initialMasterNodesFile: initialMasterNodesFile,
		endpointTypeFile:       endpointTypeFile,
	}
}

// RunRedisNodeFinder get Redis DB object and extract shard/replica count, and initial master nodes name and write them to given
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

	dnsInfo, err := r.validGivenAnnounces(db)
	if err != nil {
		internalDnsInfo := make([]string, 0)
		r.waitUntilAllPodGetItsIP(db)
		for shardNo := 0; shardNo < dbShardCount; shardNo++ {
			shardName := fmt.Sprintf("%s-shard%d", r.RedisName, shardNo)
			petset, err := r.psClient.AppsV1().PetSets(r.Namespace).Get(context.TODO(), shardName, metav1.GetOptions{})
			if err != nil {
				klog.Fatalln(err)
				return
			}
			for podNo := 0; podNo < dbReplicaCount; podNo++ {
				podName := fmt.Sprintf("%s-%d", shardName, podNo)

				pod, err := r.coreV1Client.Pods(db.Namespace).Get(context.TODO(), podName, metav1.GetOptions{})
				if err != nil {
					klog.Fatalln(err)
					return
				}
				dnsName := pod.Status.PodIP

				dbPort, dbBusPort := kubedb.RedisDatabasePort, kubedb.RedisGossipPort
				for _, container := range petset.Spec.Template.Spec.Containers {
					if container.Name != kubedb.RedisContainerName {
						continue
					}
					for _, port := range container.Ports {
						if port.Name == kubedb.RedisDatabasePortName {
							dbPort = int(port.ContainerPort)
						} else if port.Name == kubedb.RedisGossipPortName {
							dbBusPort = int(port.ContainerPort)
						}
					}
				}
				internalDnsInfo = append(internalDnsInfo, fmt.Sprintf("%s %s %d %d", podName, dnsName, dbPort, dbBusPort))
			}
		}
		dnsInfo = internalDnsInfo
	}

	r.writePodDNSToFile(r.NodesFile, dnsInfo)

	var masterNodes []string
	for _, currPodDNS := range dnsInfo {
		infos := strings.Split(currPodDNS, " ")
		podName := infos[0]
		podNum := strings.Split(podName, "-")
		if strings.Compare(podNum[len(podNum)-1], "0") == 0 {
			masterNodes = append(masterNodes, currPodDNS)
		}
	}
	r.writePodDNSToFile(r.initialMasterNodesFile, masterNodes)

	if db.Spec.Cluster == nil {
		db.Spec.Cluster = &v1.RedisClusterSpec{}
	}
	if db.Spec.Cluster.Announce == nil {
		db.Spec.Cluster.Announce = &v1.Announce{}
	}
	if db.Spec.Cluster.Announce.Type == "" {
		db.Spec.Cluster.Announce.Type = v1.PreferredEndpointTypeIP
	}
	r.writeEndpointTypeToFile(r.endpointTypeFile, db.Spec.Cluster.Announce.Type)
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

func (r *RedisdNodeFinder) writeEndpointTypeToFile(filename string, endpointType v1.PreferredEndpointType) {
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
	_, err = file.WriteString(string(endpointType))
	if err != nil {
		klog.Fatalln(err)
		return
	}
}

func (r *RedisdNodeFinder) validGivenAnnounces(rd *v1.Redis) ([]string, error) {
	if rd.Spec.Cluster == nil || rd.Spec.Cluster.Announce == nil || rd.Spec.Cluster.Announce.Shards == nil {
		return []string{}, errors.New("cluster or announce shards is empty")
	}
	announceList := rd.Spec.Cluster.Announce.Shards

	if len(announceList) != int(*rd.Spec.Cluster.Shards) {
		return []string{}, errors.New("invalid cluster or announce shards")
	}

	dnsInfo := make([]string, 0)

	for i, announceListForShard := range announceList {
		if len(announceListForShard.Endpoints) != int(*rd.Spec.Cluster.Replicas) {
			return []string{}, errors.New("invalid cluster or announce shards")
		}
		shardName := fmt.Sprintf("%s-shard%d", r.RedisName, i)
		for j, announceForReplicas := range announceListForShard.Endpoints {
			podName := fmt.Sprintf("%s-%d", shardName, j)
			hostPort := strings.Split(announceForReplicas, ":")
			host := hostPort[0]
			portBusPort := strings.Split(hostPort[1], "@")
			port := portBusPort[0]
			busPort := portBusPort[1]
			dnsInfo = append(dnsInfo, fmt.Sprintf("%s %s %s %s", podName, host, port, busPort))
		}
	}
	return dnsInfo, nil
}

func (r *RedisdNodeFinder) waitUntilAllPodGetItsIP(rd *v1.Redis) {
	assignedIpForAll := false
	dbShardCount := int(*rd.Spec.Cluster.Shards)
	dbReplicaCount := int(*rd.Spec.Cluster.Replicas)
	for !assignedIpForAll {
		assignedIpForAll = true
		for shardNo := 0; shardNo < dbShardCount; shardNo++ {
			shardName := fmt.Sprintf("%s-shard%d", r.RedisName, shardNo)
			for podNo := 0; podNo < dbReplicaCount; podNo++ {
				podName := fmt.Sprintf("%s-%d", shardName, podNo)
				pod, err := r.coreV1Client.Pods(rd.Namespace).Get(context.TODO(), podName, metav1.GetOptions{})
				if err != nil || pod.Status.PodIP == "" {
					assignedIpForAll = false
					break
				}
			}
		}
	}
}
