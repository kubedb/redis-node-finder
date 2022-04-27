[![Build Status](https://github.com/kubedb/redis-node-finder/workflows/CI/badge.svg)](https://github.com/kubedb/redis-node-finder/actions?workflow=CI)
[![Slack](http://slack.kubernetes.io/badge.svg)](http://slack.kubernetes.io/#kubedb)
[![Twitter](https://img.shields.io/twitter/follow/kubedb.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=kubedb)

# redis-node-finder

## Why redis-node-finder?
To generate dns names, we could not use `peer-finder`. Because 
- peer-finder runs on post start hook and it can not update the dns names list
- we need at least three redis nodes to create cluster. So we need to wait for other pods to be up
- for our use case, each pod should not exit from init script until it joins cluster. If current pod does not know about other pod, it can not decide what to do.
- as peer-finder runs on post start hook, we need to start the redis-server before joining cluster. So it may accept client requests before joining cluster

So, we wrote another binary `redis-node-finder`:
- It gets the db object and generates all the dns names of the redis nodes
- As we know the dns names, we can check and wait for the necessary pods to be ready to create cluster
