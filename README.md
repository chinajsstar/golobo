# Golobo 1.1
Golobo is a distributed service framework dedicated to providing high-performance and transparent RPC remote service invocation. Golobo based on ZooKeeper to store configuration, Grpc to RPC remote service and <a href="https://github.com/amamina/tunnel">tunnel</a> to **Auto** detect cluster which service node active or not, found someone node joined into cluster.
## How to use
 <a href="https://github.com/amamina/golobo/tree/master/example">example</a> show how to use,
 <a href="https://github.com/amamina/golobo/blob/master/util_test.go">util_test.go</a> show how to list/remove configuration published to zookeeper.
