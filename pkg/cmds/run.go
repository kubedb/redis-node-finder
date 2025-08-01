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

package cmds

import (
	"fmt"

	redis_finder "kubedb.dev/redis-node-finder/pkg/node-finder/redis-finder"
	sentinel_finder "kubedb.dev/redis-node-finder/pkg/node-finder/sentinel-finder"

	"github.com/spf13/cobra"
)

var (
	mode              string
	sentinelFile      string
	masterFile        string
	slaveFile         string
	nodesFile         string
	initialMasterFile string
	endpointTypeFile  string
	cmd               = &cobra.Command{
		Use:               "run",
		Short:             "Launch Redis Node Finder",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if mode == "cluster" {
				fmt.Println("Running node finder for cluster mode nodes")
				c := redis_finder.New(masterFile, slaveFile, nodesFile, initialMasterFile, endpointTypeFile)

				c.RunRedisNodeFinder()
			} else if mode == "sentinel" {
				fmt.Println("Running node finder for sentinels")
				c := sentinel_finder.New(sentinelFile)
				c.RunSentinelReplicaFinder()
			} else {
				fmt.Println("Unknown mode ", mode)
			}
		},
	}
)

func NewCmdRun() *cobra.Command {
	return cmd
}

func init() {
	cmd.PersistentFlags().StringVar(&masterFile, "master-file", "master.txt", "Contains master count")
	cmd.PersistentFlags().StringVar(&slaveFile, "slave-file", "slave.txt", "Contains slave count")
	cmd.PersistentFlags().StringVar(&nodesFile, "nodes-file", "db-nodes.txt", "Contains dns names of database nodes")
	cmd.PersistentFlags().StringVar(&initialMasterFile, "initial-master-file", "initial-master-nodes.txt", "Contains dns names of initial masters")
	cmd.PersistentFlags().StringVar(&endpointTypeFile, "endpoint-type-file", "endpoint-type.txt", "Contains preferred endpoint type")

	cmd.PersistentFlags().StringVar(&mode, "mode", "cluster", "Contains Database Mode")
	cmd.PersistentFlags().StringVar(&sentinelFile, "sentinel-file", "sentinel-replicas.txt", "Contains sentinel count")
}
