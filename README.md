# Generic Purpose of app

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
go run main.go launchHttp --appName backend --config config/app-config.yaml
```

If everything goes well, the console should show,

```text
2023-07-04T09:22:32.528+0800    INFO    cmd/sam_cmd.go:75       args    {"args": []}
2023-07-04T09:22:32.528+0800    INFO    cmd/sam_cmd.go:88       Reading configuration   {"port": 5115, "endpoints": [{"name":"google","url":"https://www.google.com"},{"name":"frontend","url":"http://frontend.magellan.svc.cluster.local:8080/ping"},{"name":"backend","url":"http://backend.magellan.svc.cluster.local:8081/ping"},{"name":"storage","url":"http://storage.magellan.svc.cluster.local:8082/ping"}]}
2023-07-04T09:22:32.528+0800    INFO    cmd/sam_cmd.go:97       Final endpoints {"filteredEndPoints": [{"name":"google","url":"https://www.google.com"},{"name":"frontend","url":"http://frontend.magellan.svc.cluster.local:8080/ping"},{"name":"storage","url":"http://storage.magellan.svc.cluster.local:8082/ping"}]}
2023-07-04T09:22:32.528+0800    INFO    cmd/sam_cmd.go:103      args    {"args": [], "params": "backend"}
2023-07-04T09:22:32.529+0800    INFO    cmd/sam_cmd.go:42       starting http server    {"appName": "backend", "address": ":5115"}
  _                      _                         _
 | |__     __ _    ___  | | __   ___   _ __     __| |
 | '_ \   / _` |  / __| | |/ /  / _ \ | '_ \   / _` |
 | |_) | | (_| | | (__  |   <  |  __/ | | | | | (_| |
 |_.__/   \__,_|  \___| |_|\_\  \___| |_| |_|  \__,_|
```

The port is 5115 by default, so that, try to make the http call by using any rest client, eg. [Postman](https://www.postman.com/downloads/) or `curl`.


```shell
curl 127.0.0.1:5115/propagate | jq
```
>In case you have [CLI json processor a.k.a jq](https://stedolan.github.io/jq/) installed

You should see the following response,

```json
[
  {
    "response_code": 200,
    "message": "200 OK",
    "from": "frontend.magellan.svc.cluster.local",
    "to": "http://backend:5115/ping",
    "dns_checking": "backend:5115 is resolved succesfully, ip address [172.20.0.3] "
  },
  {
    "response_code": 200,
    "message": "200 OK",
    "from": "frontend.magellan.svc.cluster.local",
    "to": "https://www.google.com",
    "dns_checking": "www.google.com is resolved succesfully, ip address [142.251.12.147 142.251.12.99 142.251.12.104 142.251.12.105 142.251.12.106 142.251.12.103 2404:6800:4003:c1a::67 2404:6800:4003:c1a::93 2404:6800:4003:c1a::63 2404:6800:4003:c1a::69] "
  }
]
```

All right, you are good to continue with development !!.

## Containerization
The first step to make the application available for cloud deployment, is to make it a container image. For this, some tools are available, eg. [docker](https://www.docker.com/). What it is needed is Dockerfile in the root folder of this project. The following step requires the docker installed in your local machine.

The command to build the docker image is,

```shell
docker build -t samutup/http-ping:0.0.8 -f Dockerfile .
```
If everything goes well you should see,

```text
...
 => exporting to image                                                                                                      0.1s
 => => exporting layers                                                                                                     0.1s
 => => writing image sha256:e563410b71a23c20cf241cea94b126453883b8ff4268b50a8b864d0130334c08                                0.0s
 => => naming to docker.io/samutup/http-ping:0.0.8
```

All right, now a brand new docker image is created in the local docker registry. 
The next step is to run it. With the help of `docker-compose`, 
running the docker container from an image is more convenient. 
There is docker compose file, in this project under, `sam-ping-compose` folder.

This file, defines the the service configuration for the container to be run.

```yaml
---
version: '2'
services:
  frontend:
    image: samutup/http-ping:0.0.8
    command: ["/app/http-ping","launchHttp","--appName=frontend","--config=/app/config/sam-ping-docker.yaml" ]
    hostname: frontend.magellan.svc.cluster.local
    container_name: frontend
    ports:
      - "8080:5115"
    volumes:
      - "./config:/app/config"
  backend:
    image: samutup/http-ping:0.0.8
    command: ["/app/http-ping","launchHttp","--appName=backend","--config=/app/config/sam-ping-docker.yaml" ]
    hostname: backend.magellan.svc.cluster.local
    container_name: backend
    ports:
      - "8081:5115"
    volumes:
      - "./config:/app/config"
```
In this case, 2 containers will run with the name, `frontend` and `backend`
It exposes the port `8080` for frontend and `8081` for `backend`. Both are listening at port `5115`
To reach the `backend` from `frontend` within the container runtime, it is simply, `GET http://backend:5115/ping`,
while, to reach the backend from, outside of the container runtime, eg, from the `host` machine (developer machine),
`GET http://127.0.0.1:8080/ping`. For this gets to work the configuration file for docker environment looks like,

```yaml
port: 5115
endPoints:
- name: google
  url: https://www.google.com
- name: frontend
  url: http://frontend:5115/ping
- name: backend
  url: http://backend:5115/ping
```

Please take a note that, through `volumes` directive, we override the existing `/app/config/sam-ping.yaml` file with the `./config/sam-ping-docker.yaml` from the host's folder. That is the reason why the docker compose overrides the `ENTRYPOINT` docker file instruction to be, `/app/http-ping","launchHttp","--appName=frontend","--config=/app/config/sam-ping-docker.yaml` through `command` directive.


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
> docker-compose logs --follow
backend   | 2023-07-04T02:55:01.743Z    INFO    cmd/sam_cmd.go:75       args    {"args": ["/app/http-ping", "launchHttp"]}
backend   | 2023-07-04T02:55:01.746Z    INFO    cmd/sam_cmd.go:88       Reading configuration   {"filename": "/app/config/sam-ping-docker.yaml", "port": 5115, "endpoints": [{"name":"google","url":"https://www.google.com"},{"name":"frontend","url":"http://frontend:5115/ping"},{"name":"backend","url":"http://backend:5115/ping"}]}
backend   | 2023-07-04T02:55:01.746Z    INFO    cmd/sam_cmd.go:97       Final endpoints {"filteredEndPoints": [{"name":"google","url":"https://www.google.com"},{"name":"frontend","url":"http://frontend:5115/ping"}]}
backend   | 2023-07-04T02:55:01.746Z    INFO    cmd/sam_cmd.go:103      args    {"args": ["/app/http-ping", "launchHttp"], "params": "backend"}
backend   | 2023-07-04T02:55:01.747Z    INFO    cmd/sam_cmd.go:42       starting http server    {"appName": "backend", "address": ":5115"}
backend   |   _                      _                         _
backend   |  | |__     __ _    ___  | | __   ___   _ __     __| |
backend   |  | '_ \   / _` |  / __| | |/ /  / _ \ | '_ \   / _` |
backend   |  | |_) | | (_| | | (__  |   <  |  __/ | | | | | (_| |
backend   |  |_.__/   \__,_|  \___| |_|\_\  \___| |_| |_|  \__,_|
frontend  | 2023-07-04T02:55:01.718Z    INFO    cmd/sam_cmd.go:75       args    {"args": ["/app/http-ping", "launchHttp"]}
frontend  | 2023-07-04T02:55:01.722Z    INFO    cmd/sam_cmd.go:88       Reading configuration   {"filename": "/app/config/sam-ping-docker.yaml", "port": 5115, "endpoints": [{"name":"google","url":"https://www.google.com"},{"name":"frontend","url":"http://frontend:5115/ping"},{"name":"backend","url":"http://backend:5115/ping"}]}
frontend  | 2023-07-04T02:55:01.722Z    INFO    cmd/sam_cmd.go:97       Final endpoints {"filteredEndPoints": [{"name":"google","url":"https://www.google.com"},{"name":"backend","url":"http://backend:5115/ping"}]}
frontend  | 2023-07-04T02:55:01.722Z    INFO    cmd/sam_cmd.go:103      args    {"args": ["/app/http-ping", "launchHttp"], "params": "frontend"}
frontend  | 2023-07-04T02:55:01.722Z    INFO    cmd/sam_cmd.go:42       starting http server    {"appName": "frontend", "address": ":5115"}
frontend  |    __                          _                        _
frontend  |   / _|  _ __    ___    _ __   | |_    ___   _ __     __| |
frontend  |  | |_  | '__|  / _ \  | '_ \  | __|  / _ \ | '_ \   / _` |
frontend  |  |  _| | |    | (_) | | | | | | |_  |  __/ | | | | | (_| |
frontend  |  |_|   |_|     \___/  |_| |_|  \__|  \___| |_| |_|  \__,_|
frontend  | 2023-07-04T02:55:48.189Z    INFO    cmd/handler.go:120      Serving request {"origin": "127.0.0.1:8080"}
frontend  | 2023-07-04T02:55:48.190Z    INFO    cmd/handler.go:87       ip address is resolved properly {"host": "backend:5115", "ip address": ["172.20.0.3"]}
```

Now, it is time to test it,

```shell
curl 127.0.0.1:8080/propagate | jq
```

The following is the expected output,

```json
[
  {
    "response_code": 200,
    "message": "200 OK",
    "from": "frontend.magellan.svc.cluster.local",
    "to": "http://backend:5115/ping",
    "dns_checking": "backend:5115 is resolved succesfully, ip address [172.20.0.3] "
  },
  {
    "response_code": 200,
    "message": "200 OK",
    "from": "frontend.magellan.svc.cluster.local",
    "to": "https://www.google.com",
    "dns_checking": "www.google.com is resolved succesfully, ip address [142.251.12.147 142.251.12.99 142.251.12.104 142.251.12.105 142.251.12.106 142.251.12.103 2404:6800:4003:c1a::67 2404:6800:4003:c1a::93 2404:6800:4003:c1a::63 2404:6800:4003:c1a::69] "
  }
]
```
To deploy into kubernetes cluster, please follow, [Deploy to kubernetes](./docs/deploy.md)








