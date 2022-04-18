package main

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/clientcmd"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
)

var (
	err error
	//hostName   string
	kubeConfig *restclient.Config
	//kubeClient *kubernetes.Clientset
	dbClient     *cs.Clientset
	offShootName string
)

func init() {
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
	fmt.Println("Finding redis nodes")
	printDNSOfAllRedisNodes()
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

// We get db object and check how many masters and replicas are there
// Then we generate dns names of the pods using pod name and governing service name
// Then all the infos are stored in files in /tmp folder so that we can access from bash script
func printDNSOfAllRedisNodes() {
	nameSpace := os.Getenv("NAMESPACE")
	offShootName = os.Getenv("OFF_SHOOT_NAME")
	dbGoverningServiceName := os.Getenv("REDIS_GOVERNING_SERVICE")

	db, err := dbClient.KubedbV1alpha2().Redises(nameSpace).Get(context.TODO(), offShootName, metav1.GetOptions{})
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
		shardName := fmt.Sprintf("%s-shard%d", offShootName, shardNo)

		for podNo := 0; podNo <= dbReplicaCount; podNo++ {
			podName := fmt.Sprintf("%s-%d", shardName, podNo)
			dnsName := podName + "." + dbGoverningServiceName
			redisNodes = append(redisNodes, dnsName)
		}
	}
	writePodDNSToFile("redis-nodes.txt", redisNodes)

	var masterNodes []string
	for shardNO := 0; shardNO < dbMasterCount; shardNO++ {
		initialMasterPod := fmt.Sprintf("%s-shard%d-0", offShootName, shardNO)
		dnsName := initialMasterPod + "." + dbGoverningServiceName
		masterNodes = append(masterNodes, dnsName)
	}
	writePodDNSToFile("initial-master-nodes.txt", masterNodes)
}
