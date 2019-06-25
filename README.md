Go Bench Suite
==============

Toolkit for benchmarking & tuning proxies & gateways against a mock upstream server.

```text
Step 1:                                                                      
┌────────────────┐─────────────1k rps──────────────▶┌────────────────┐
│ Load Generator │                                  │ Go Bench Suite │
└────────────────┘◀─────────────1 ms────────────────└────────────────┘
                                                                      
Step 2:                                                                      
┌────────────────┐─1k rps─▶┌───────────────┐───────▶┌────────────────┐
│ Load Generator │         │ Proxy to test │        │ Go Bench Suite │
└────────────────┘◀──5 ms──└───────────────┘◀───────└────────────────┘  

Step 3: Tuning

Step 4:
┌────────────────┐─1k rps─▶┌───────────────┐───────▶┌────────────────┐
│ Load Generator │         │ Proxy to test │        │ Go Bench Suite │
└────────────────┘◀──2 ms──└───────────────┘◀───────└────────────────┘

Step 5: Grab Coffee                                                           
```

## API Endpoints

### `GET /delay/{duration}`

Returns a 0 length body response after a `{duration}` delay. 
Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

```bash
date && curl localhost:8000/delay/5s -i && date
Sat Jun  1 19:02:23 BST 2019
HTTP/1.1 200 OK
Server: fasthttp
Date: Sat, 01 Jun 2019 18:02:27 GMT
Content-Length: 0

Sat Jun  1 19:02:28 BST 2019
```

### `GET /json/{type}`

Return JSON response body of type `invalid` or `valid`.

```bash
curl localhost:8000/json/valid -i      
HTTP/1.1 200 OK
Server: fasthttp
Date: Sat, 01 Jun 2019 18:04:32 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 65

{"time": "2019-06-01 19:04:32.592929 +0100 BST m=+242.881838180"}
```

```bash
curl localhost:8000/json/invalid -i
HTTP/1.1 200 OK
Server: fasthttp
Date: Sat, 01 Jun 2019 18:04:53 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 64

{time": "2019-06-01 19:04:53.364517 +0100 BST m=+263.653200176"}
```

We can also simulate upstream latency by forcing a synthetic delay using `X-Delay: {delay}` in the request header.

```bash
date && curl localhost:8000/json/valid -H 'X-Delay: 2s' -i && date
HTTP/1.1 200 OK
Server: fasthttp
Date: Sat, 01 Jun 2019 18:04:32 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 65

{"time": "2019-06-01 19:04:32.592929 +0100 BST m=+242.881838180"}
Sat Jun  1 19:15:47 BST 2019
```

### `GET /xml`

Returns an XML response body

```bash
curl localhost:8000/xml -i         
HTTP/1.1 200 OK
Server: fasthttp
Date: Sat, 01 Jun 2019 18:06:58 GMT
Content-Type: application/xml
Content-Length: 470

<?xml version='1.0' encoding='us-ascii'?>
<!--  A SAMPLE set of slides  -->
<slideshow title="Sample Slide Show" date="Date of publication" author="Yours Truly">
  <!-- TITLE SLIDE -->
  <slide type="all">
    <title>Wake up to WonderWidgets!</title>
  </slide>

  <!-- OVERVIEW -->
  <slide type="all">
    <title>Overview</title>
    <item>Why <em>WonderWidgets</em> are great</item>
    <item/>
    <item>Who <em>buys</em> WonderWidgets</item>
  </slide>
</slideshow>
```

### `GET|POST|PUT|PATCH /size/{size}`

Return random chars of size `{size}`.
Size accepts a human readable byte string into a number of bytes. 
Supports `B`, `K`, `KB`, `M`, `MB`.

```bash
curl localhost:8000/size/0.25K -i
HTTP/1.1 200 OK
Server: fasthttp
Date: Sat, 01 Jun 2019 18:13:52 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 256

lIzTIYPnAufxohekhiaQtnkYHUOLmYMrHMBaCtkdVtjoRCbNHsXeZsdtvQMNWOAxoTWwACfbmVAQKtVEsWgeSKZhIWYxONgTAdOPOZbjkjRNdikyBwsrIzSldLpedgNqouebcgVhVzhzPNacRuLTCciGetmEyWkdgnwzoJZwAnmXfajOwPlCPOSDHVcKYqEZxvkLixeXGMamyEikqugebsCEKSwmbQvwUWnNSwfpLHqUIfkdOcgkDCmbTvAypNlX
```

We can also simulate upstream latency by forcing a synthetic delay using `X-Delay: {delay}` in the request header.

```bash
date && curl localhost:8000/size/10B -H 'X-Delay: 2s' -i && date
Sat Jun  1 19:15:45 BST 2019
HTTP/1.1 200 OK
Server: fasthttp
Date: Sat, 01 Jun 2019 18:15:46 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 10

EAYWJgKmOG
Sat Jun  1 19:15:47 BST 2019
```

Or POST a 5KB request and return a 20B response

```bash
curl -X POST -d "$(curl -s -X GET localhost:8095/size/5KB)" localhost:8095/size/20B -i
HTTP/1.1 100 Continue

HTTP/1.1 200 OK
Server: fasthttp
Date: Sat, 01 Jun 2019 18:19:35 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 20

hPfepqqojzMLQIRYPnnr
```

### `GET /resource`

Returns some resources, default 10, but optional `limit` query string parameter.
This also accepts `X-Delay` query string parameter

```
curl http://localhost:8000/resource\?limit\=5
{"0":{"Id":0,"Name":"aJjLLZRjCV"},"1":{"Id":1,"Name":"IOrvpPwJkO"},"2":{"Id":2,"Name":"kfakGCrQIn"},"3":{"Id":3,"Name":"prDmIjElNj"},"4":{"Id":4,"Name":"bbGzvRrJMT"}}
```

### `GET /resource/{id}`

Returns a particular resource by it's id

```bash
curl http://localhost:8095/resource/3 -H 'X-Delay: 3s'
{"Id":3,"Name":"prDmIjElNj"}%
```

## Chaos & Mocking a SLOW response

It is possible to introduce a delay from any endpoint by setting `X-Delay` header in the http request.
But let's assume that we don't want the response to be slow / delayed ALL the time. We can also set an `X-Delay-Percent`
header in the request, which takes an integer percentage value.

Examples:

`curl http://localhost:8095/resource/3 -H 'X-Delay: 3s'` returns response in ~3s ALL the time
`curl http://localhost:8095/resource/3 -H 'X-Delay: 3s' -H 'X-Delay-Percent: 100'` returns response in ~3s ALL the time
`curl http://localhost:8095/resource/3 -H 'X-Delay: 3s' -H 'X-Delay-Percent: 5'` returns response in ~3s 5% of the time, otherwise, response is immediate.

## Installation

### From Source

```bash
go install && go-bench-suite upstream

# or to change listen address
go install && go-bench-suite upstream --addr 127.0.0.1:8000
```

### From Source + TLS

```bash
go install && go-bench-suite upstream --addr 127.0.0.1:8443 --certFile ./certs/foo.com.cert.pem --keyFile ./certs/foo.com.key.pem
```

### Docker

```bash
docker network create bench
docker run --rm -itd --name bench --network bench mangomm/go-bench-suite ./go-bench-suite upstream
docker run --rm -it --network bench rcmorano/docker-hey -z 30s http://bench:8000/size/10B
```

### Kubernetes & Load testing

Create our mock upstream server deployment & service

```bash
kubectl apply -f ./k8s/upstream

namespace "upstream" created
deployment.apps "upstream" created
namespace "upstream" configured
service "upstream" created
```

Check it's working

```bash
# terminal 1

kubectl port-forward --namespace upstream svc/upstream 8000:8000

Forwarding from 127.0.0.1:8000 -> 8000
Forwarding from [::1]:8000 -> 8000
Handling connection for 8000
```

```bash
# terminal 2

curl localhost:8000/size/10B -i

HTTP/1.1 200 OK
Server: fasthttp
Date: Fri, 08 Feb 2019 21:19:14 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 10

dZapAiUvaE
```

Benchmark upstream

```bash
kubectl run foo --rm -i -t --image=rcmorano/docker-hey -- -z 10s -H 'X-Delay: 10ms' http://upstream.upstream:8081/size/10B

Summary:
  Total:        10.0107 secs
  Slowest:		0.0452 secs
  Fastest:		0.0102 secs
  Average:		0.0112 secs
  Requests/sec:	4477.0918

--- snip ---
```
