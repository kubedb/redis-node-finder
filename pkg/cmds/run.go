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

package cmds

import (
	"github.com/Shaad7/redis-node-finder/pkg/node_finder"
	"github.com/spf13/cobra"
	"io"
)

func NewCmdRun(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Launch Redis Node Finder",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			c := node_finder.New()
			c.RunRedisCoordinator(stopCh)
		},
	}

	return cmd
}
