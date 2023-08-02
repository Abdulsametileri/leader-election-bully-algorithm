# Working Demo

[![asciicast](https://asciinema.org/a/600162.svg)](https://asciinema.org/a/600162)

# Quickstart with Docker

This project has an already configured [Docker Compose file](docker-compose.yml) launching four nodes.

You can build and run docker-compose as follows:

`docker-compose up --build`

You can kill some nodes with docker commands to test the cluster behavior.

If you want to add new nodes, please add its address to hardcoded variables.

```go
// nodeAddressByID: It includes nodes currently in cluster
var nodeAddressByID = map[string]string{
	"node-01": "node-01:6001",
	"node-02": "node-02:6002",
	"node-03": "node-03:6003",
	"node-04": "node-04:6004",
}
```

In this repository, I wanted to avoid implementing a service discovery mechanism. 
I aim to learn and implement basic how to do leader election without service discovery. 

You can also find [my service discovery implementation](https://github.com/Abdulsametileri/simple-service-discovery) and 
[article](https://itnext.io/lets-implement-basic-service-discovery-using-go-d91c513883f6)

# Quickstart without Docker

You can still execute this project without using Docker.

First, change the Node IPs. For example

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

## Additional ToDos

- [ ] 1- Implement with service discovery
- [ ] 2- Change ordered election to dynamic.
