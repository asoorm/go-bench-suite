Go Bench Suite
==============

Toolkit for benchmarking & testing proxies against a mock upstream server.

## API Upstream Examples

Return response in 5s
```
GET /delay/5s
```

Return JSON response body
```
GET /json/valid
GET /json/invalid // returns invalid body
```

Return XML response body
```
GET /xml
```

Return random chars of size X
```
GET /size/1MB
GET /size/1KB

# send 5KB, return 10KB
curl -X POST -d "$(curl -X GET localhost:8000/size/5KB)" localhost:8000/size/10KB

# simulate upstream latency with a delay
curl -X GET -H 'X-Delay 100ms' localhost:8000/size/10KB
``` 

Return Soap response
```
TODO
```

## Native

```
go install && go-bench-suite upstream

# or to change listen address
go install && go-bench-suite upstream --addr 127.0.0.1:8000
```

## Native + TLS

```
go install && go-bench-suite upstream --addr 127.0.0.1:8443 --certFile ./certs/foo.com.cert.pem --keyFile ./certs/foo.com.key.pem
```

## Docker

```
docker network create bench
docker run --rm -itd --name bench --network bench mangomm/go-bench-suite ./go-bench-suite upstream
docker run --rm -it --network bench rcmorano/docker-hey -z 30s http://bench:8000/size/10B
```

## Kubernetes

### upstream server

Create our mock upstream server deployment & service

```
$ kubectl apply -f ./k8s/upstream

namespace "upstream" created
deployment.apps "upstream" created
namespace "upstream" configured
service "upstream" created
```

Check it's working

```
# terminal 1

$ kubectl port-forward --namespace upstream svc/upstream 8000:8000

Forwarding from 127.0.0.1:8000 -> 8000
Forwarding from [::1]:8000 -> 8000
Handling connection for 8000
```

```
# terminal 2

$ curl localhost:8000/size/10B -i

HTTP/1.1 200 OK
Server: fasthttp
Date: Fri, 08 Feb 2019 21:19:14 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 10

dZapAiUvaE
```

---

## benchmark upstream

```
$ kubectl run foo --rm -i -t --image=rcmorano/docker-hey -- -z 10s -H 'X-Delay: 10ms' http://upstream.upstream:8081/size/10B
If you don't see a command prompt, try pressing enter.

Summary:
  Total:        10.0107 secs
  Slowest:		0.0452 secs
  Fastest:		0.0102 secs
  Average:		0.0112 secs
  Requests/sec:	4477.0918

--- snip ---
```
