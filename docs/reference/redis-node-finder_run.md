## redis-node-finder run

Launch Redis Node Finder

```
redis-node-finder run [flags]
```

### Options

```
  -h, --help                         help for run
      --initial-master-file string   Contains dns names of initial masters (default "initial-master-nodes.txt")
      --master-file string           Contains master count (default "master.txt")
      --redis-nodes-file string      Contains dns names of redis nodes (default "redis-nodes.txt")
      --slave-file string            Contains slave count (default "slave.txt")
```

### Options inherited from parent commands

```
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
```

### SEE ALSO

* [redis-node-finder](redis-node-finder.md)	 - 

