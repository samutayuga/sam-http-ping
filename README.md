# sam-http-ping

## Environment

`go1.20.4`, [Visit](https://go.dev/dl/)

## Run locally

Clone the project, git clone https://github.com/samutayuga/sam-http-ping.git, then cd sam-http-ping.

Resolve the dependencies,

```shell
go mod tidy
```

Run the main file,

```shell
go run main.go
```

If everything goes well, the console should show,

```text
http-ping  | 2023-05-25T12:54:04.402Z   INFO    cmd/handler.go:135      Reading configuration   {"port": 5115, "endpoints": [{"name":"google","url":"https://www.google.com"},{"name":"frontend","url":"http://frontend.magellan.svc.cluster.local:8080"},{"name":"backend","url":"http://backend.magellan.svc.cluster.local:8081"},{"name":"storage","url":"http://storage.magellan.svc.cluster.local:8082"}]}
```

The port is 5115 by default, so that, try to make the http call by using any rest client, eg. [Postman](https://www.postman.com/downloads/) or `curl`.


```shell
curl 127.0.0.1:5115/propagate | jq
```
>In case you have [CLI json processor a.k.a jq](https://stedolan.github.io/jq/) installed

You should see the following response,

```json
...
[
  {
    "ResponseCode": 200,
    "ResponseMessage": "200 OK",
    "Origin": "http-ping",
    "Destination": "https://www.google.com"
  },
  {
    "ResponseCode": -1,
    "ResponseMessage": "Get \"http://frontend.magellan.svc.cluster.local:8080\": dial tcp: lookup frontend.magellan.svc.cluster.local: Try again",
    "Origin": "http-ping",
    "Destination": "http://frontend.magellan.svc.cluster.local:8080"
  },
  {
    "ResponseCode": -1,
    "ResponseMessage": "Get \"http://backend.magellan.svc.cluster.local:8081\": dial tcp: lookup backend.magellan.svc.cluster.local: Try again",
    "Origin": "http-ping",
    "Destination": "http://backend.magellan.svc.cluster.local:8081"
  },
  {
    "ResponseCode": -1,
    "ResponseMessage": "Get \"http://storage.magellan.svc.cluster.local:8082\": dial tcp: lookup storage.magellan.svc.cluster.local: Try again",
    "Origin": "http-ping",
    "Destination": "http://storage.magellan.svc.cluster.local:8082"
  }
]
```

All right, you are good to continue with development !!.

## Containerization
The first step to make the application available for cloud deployment, is to make it a container image. For this, some tools are available, eg. [docker](https://www.docker.com/). What it is needed is Dockerfile in the root folder of this project. The following step requires the docker installed in your local machine.

The command to build the docker image is,

```shell
docker build -t samutup/http-ping:0.0.1-SNAPSHOT -f Dockerfile .
```
If everything goes well you should see,

```text
...
 => exporting to image                                                                                                      0.1s
 => => exporting layers                                                                                                     0.1s
 => => writing image sha256:e563410b71a23c20cf241cea94b126453883b8ff4268b50a8b864d0130334c08                                0.0s
 => => naming to docker.io/samutup/http-ping:0.0.1-SNAPSHOT 
```

All right, now a brand new docker image is created in the local docker registry. 
The next step is to run it. With the help of `docker-compose`, 
running the docker container from an image is more convenient. 
There is docker compose file, in this project under, `sam-ping-compose` folder.

This file, defines the the service configuration for the container to be run.

```yaml
version: '2'
services:
  http-ping:
    image: samutup/http-ping:0.0.1-SNAPSHOT
    hostname: http-ping.backend
    container_name: http-ping
    ports:
      - "5115:5115"
    environment:
      LABEL_ENV: backend
```
In this case, the container will run with the name, `http-ping`, with hostname, `http-ping.backend`
It exposes the port `5115` and listening at port `5115`
Go into the folder then run the command,

```shell
docker-compose up -d
```
If everything goes well you should see,

```text
[+] Running 1/1
 ⠿ Container http-ping  Started 
```

Tail the container log,

```shell
docker-compose logs --follow
```

You should see,

```text
export APP_NAME=BACKEND && go run main.go
2023-06-29T08:51:40.306+0800    INFO    cmd/handler.go:163      Reading configuration   {"port": 5115, "endpoints": [{"name":"google","url":"https://www.google.com"},{"name":"frontend","url":"http://frontend.magellan.svc.cluster.local:8080/ping"},{"name":"backend","url":"http://backend.magellan.svc.cluster.local:8081/ping"},{"name":"storage","url":"http://storage.magellan.svc.cluster.local:8082/ping"}]}
2023-06-29T08:51:40.306+0800    INFO    sam-http-ping/main.go:26        starting http server    {"appName": "BACKEND", "address": ":5115"}
  ____       _       ____   _  __  _____   _   _   ____
 | __ )     / \     / ___| | |/ / | ____| | \ | | |  _ \
 |  _ \    / _ \   | |     | ' /  |  _|   |  \| | | | | |
 | |_) |  / ___ \  | |___  | . \  | |___  | |\  | | |_| |
 |____/  /_/   \_\  \____| |_|\_\ |_____| |_| \_| |____/
```

To deploy into kubernetes cluster, please follow, [Deploy to kubernetes](./docs/deploy.md)








