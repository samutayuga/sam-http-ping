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
* Connected to 127.0.0.1 (127.0.0.1) port 5115 (#0)
> GET /propagate HTTP/1.1
> Host: 127.0.0.1:5115
> User-Agent: curl/7.77.0
> Accept: */*
> 
  0     0    0     0    0     0      0      0 --:--:--  0:00:06 --:--:--     0* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Thu, 25 May 2023 12:54:26 GMT
< Content-Length: 841
< Content-Type: text/plain; charset=utf-8
< 
{ [841 bytes data]
100   841  100   841    0     0    119      0  0:00:07  0:00:07 --:--:--   175
* Connection #0 to host 127.0.0.1 left intact
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

All right, now a brand new docker image is created in the local docker registry. The next step is to run it. With the helm of `docker-compose`, running the docker container from an image is more abstract. There is docker compose file, in this project under, `sam-ping-compose` folder.
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
http-ping  | 2023-05-25T12:54:04.402Z   INFO    cmd/handler.go:135      Reading configuration   {"port": 5115, "endpoints": [{"name":"google","url":"https://www.google.com"},{"name":"frontend","url":"http://frontend.magellan.svc.cluster.local:8080"},{"name":"backend","url":"http://backend.magellan.svc.cluster.local:8081"},{"name":"storage","url":"http://storage.magellan.svc.cluster.local:8082"}]}
http-ping  | 2023-05-25T12:54:04.402Z   INFO    app/main.go:22  starting http server    {"address": ":5115"}
```

All right, that's it.









