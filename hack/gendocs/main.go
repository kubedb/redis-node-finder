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

package main

import (
	"fmt"
	"log"
	"os"

	"kubedb.dev/redis-node-finder/pkg/cmds"

	"github.com/spf13/cobra/doc"
	"gomodules.xyz/runtime"
)

func docsDir() string {
	if dir, ok := os.LookupEnv("DOCS_ROOT"); ok {
		return dir
	}
	return runtime.GOPath() + "/src/kubedb.dev/redis-node-finder"
}

// ref: https://github.com/spf13/cobra/blob/master/doc/md_docs.md
func main() {
	rootCmd := cmds.NewRootCmd()
	dir := docsDir() + "/docs/reference"
	fmt.Printf("Generating cli markdown tree in: %v\n", dir)
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(dir, 0o755)
	if err != nil {
		log.Fatal(err)
	}
	_ = doc.GenMarkdownTree(rootCmd, dir)
}
