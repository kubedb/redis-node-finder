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
	node_finder "kubedb.dev/redis-node-finder/pkg/node-finder"

	"github.com/spf13/cobra"
)

var (
	masterFile        string
	slaveFile         string
	redisNodesFile    string
	initialMasterFile string
	cmd               = &cobra.Command{
		Use:               "run",
		Short:             "Launch Redis Node Finder",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			c := node_finder.New(masterFile, slaveFile, redisNodesFile, initialMasterFile)
			c.RunRedisNodeFinder()
		},
	}
)

func NewCmdRun() *cobra.Command {
	return cmd
}

func init() {
	cmd.PersistentFlags().StringVar(&masterFile, "master-file", "master.txt", "Contains master count")
	cmd.PersistentFlags().StringVar(&slaveFile, "slave-file", "slave.txt", "Contains slave count")
	cmd.PersistentFlags().StringVar(&redisNodesFile, "redis-nodes-file", "redis-nodes.txt", "Contains dns names of redis nodes")
	cmd.PersistentFlags().StringVar(&initialMasterFile, "initial-master-file", "initial-master-nodes.txt", "Contains dns names of initial masters")
}
