# Introduction

When I read a book called [Database Internals](https://www.amazon.com/Database-Internals-Deep-Distributed-Systems/dp/1492040347) 
I read a chapter about leader election. During the reading in order to learn more
I decided to implement the simplest algorithm called *bully algorithm*. This repository
is an implementation of [bully algorithm](https://en.wikipedia.org/wiki/Bully_algorithm) 
in Go using [Remote Produce Call (RPC)](https://pkg.go.dev/net/rpc).

In the bully algorithm the fundamental idea is rank. It assumes that every node
has a rank within cluster and the leader must be the highest one. So it uses node's
rank value during the election.

There are two situations for election.  
* System is newly initialized so there is no leader
* One of the nodes notices that the leader is down.

Election was implemented as follows:
1. The node send "**ELECTION**" messages to the other nodes which has higher rank that own rank.
2. The node waits for "**ALIVE**" responses.
- If no higher-ranked node responds, it makes itself a leader.
- Otherwise, it is notified the new leader which has the highest rank.

Let's illustrate these scenarios: 

We assumed that the highest rank order like: **node-04 > node-03 > node-02 > node-01**

- If the system is newly initialized 

![election step one](assets/election-step-1.png)

Because of node-04 didn't get alive message, it makes itself a leader, and it broadcasts
"**ELECTED**" message for notifying other nodes about the election results.

![election step two](assets/election-step-2.png)

In order to notice the leader is down, the other nodes periodically send "**PING**" messages
and waiting the leader "**PONG**" responses. It the leader is down and the first node didn't get
"**PONG**" message, that node starts the election process again.

![ping pong step](assets/ping-pong-step.png)

For example, node-01 didn't get PONG response from leader, it starts the election
process again. Same processes as shown above is applied and node-03 will be a new leader

![new leader](assets/new-leader.png)

# Working Demo

[![asciicast](https://asciinema.org/a/600162.svg)](https://asciinema.org/a/600162)

# Quickstart with Docker

This project comes with an already configured [Docker Compose file](docker-compose.yml) launching four nodes.

You can build and run docker-compose as follows:

`docker-compose up --build`

If you want to test the cluster behaviour, you can kill some of the nodes with docker commands.

If you want to add new nodes, please add its address to hardcoded variables

```go
// nodeAddressByID: It includes nodes currently in cluster
var nodeAddressByID = map[string]string{
	"node-01": "node-01:6001",
	"node-02": "node-02:6002",
	"node-03": "node-03:6003",
	"node-04": "node-04:6004",
}
```

In this repository, I don't want to implement service discovery mechanism. My aim is to learn
and implement basic how to do leader election :) You can find my service discovery implementation and article 
[there](https://github.com/Abdulsametileri/simple-service-discovery) to learn more about it! 

# Quickstart without Docker

You can still execute this project without using Docker.

First, change the node ips. For example

```go
// nodeAddressByID: It includes nodes currently in cluster
var nodeAddressByID = map[string]string{
	"node-01": "node-01:6001", <--- change here to localhost:6001
	"node-02": "node-02:6002", <--- change here to localhost:6002
	"node-03": "node-03:6003", <--- change here to localhost:6003
	"node-04": "node-04:6004", <--- change here to localhost:6004
}
```

Second, you can set which node you want to run in `Program arguments` like

`go run . node-02`

