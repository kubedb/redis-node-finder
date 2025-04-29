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

package sentinel_finder

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

type SentinelReplicaFinder struct {
	Namespace      string
	dbClient       *cs.Clientset
	sentinelDBName string
	sentinelFile   string
}

func New(sentinelFile string) *SentinelReplicaFinder {
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		klog.Fatalln(err)
	}
	clientcmd.Fix(kubeConfig)
	dbClient, err := cs.NewForConfig(kubeConfig)
	if err != nil {
		klog.Fatalln(err)
	}
	namespace := os.Getenv("SENTINEL_NAMESPACE")
	SentinelName := os.Getenv("SENTINEL_NAME")

	return &SentinelReplicaFinder{
		dbClient:       dbClient,
		Namespace:      namespace,
		sentinelDBName: SentinelName,
		sentinelFile:   sentinelFile,
	}
}

// RunSentinelReplicaFinder  get Redis DB  object and extract sentinel replica count and write it to given
// file name in /tmp directory. The call is made from init script, so it will write to /tmp/ directory
// The init script then use those value to provision the db object with right configuration and the init
// script also has updated information during pod restart
func (r *SentinelReplicaFinder) RunSentinelReplicaFinder() {
	db, err := r.dbClient.KubedbV1().RedisSentinels(r.Namespace).Get(context.TODO(), r.sentinelDBName, metav1.GetOptions{})
	if err != nil {
		klog.Fatalln(err)
		return
	}

	dbReplicaCount := int(*db.Spec.Replicas)

	r.writeInfoToFile(r.sentinelFile, dbReplicaCount)
}

func (r *SentinelReplicaFinder) writeInfoToFile(filename string, count int) {
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
