package upstream

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/gommon/bytes"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttprouter"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits

	// src: httpbin.org/xml
	xml = `<?xml version='1.0' encoding='us-ascii'?>
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
</slideshow>`
)

type resource struct {
	Id   int
	Name string
}

var resources map[int]resource

func Serve(addr string) error {
	logrus.Infof("starting server on %s", addr)

	r := registerRoutes()

	return fasthttp.ListenAndServe(addr, r.Handler)
}

func ServeTLS(addr, certFile, keyFile string) error {
	logrus.Infof("starting TLS server on %s", addr)

	r := registerRoutes()

	return fasthttp.ListenAndServeTLS(addr, certFile, keyFile, r.Handler)
}

func registerRoutes() *fasthttprouter.Router {
	r := fasthttprouter.New()

	r.GET("/delay/:delay", fixedDelayResponse)
	r.GET("/json/:type", jsonHandler)

	log.Printf("GET /json")
	log.Printf("\tPATH /valid\treturns a valid json response\n")
	log.Printf("\tPATH /invalid\t returns an invalid json response\n")
	log.Printf("\tHEADER X-Delay: 200ms")
	log.Printf("\t\t-> responds in 200ms")
	log.Printf("\tHEADER X-Delay: 100ms, X-Slowdown: 300ms, X-Slowdown-From: 2006-01-02T15:04:05Z07:00\n")
	log.Printf("\t\t-> responds in 200ms, from %s applies a further 300ms slowdown (time.RFC3339)",
		time.Now().Add(time.Minute).Format(time.RFC3339))

	r.GET("/xml", xmlHandler)
	r.POST("/soap", soapHandler)
	r.GET("/size/:size", sizeHandler)

	seedResources()
	r.GET("/resource", resourceIndexHandler)
	r.GET("/resource/:id", resourceShowHandler)

	return r
}

func seedResources() {
	resources = make(map[int]resource, 100)
	for i := 0; i < 100; i++ {
		resources[i] = resource{Id: i, Name: randStringBytesMaskImprSrc(10, rand.NewSource(time.Now().UnixNano()))}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func resourceIndexHandler(c *fasthttp.RequestCtx, _ fasthttprouter.Params) {
	c.SetContentType("application/json")

	if err := applyDelay(c, nil); err != nil {
		return
	}

	limit := c.QueryArgs().GetUintOrZero("limit")
	if limit == 0 {
		limit = 10
	}
	subset := make(map[int]resource)
	for i := 0; i < min(limit, len(resources)); i++ {
		subset[i] = resources[i]
	}

	jsBytes, _ := json.Marshal(subset)

	fmt.Fprint(c, string(jsBytes))

	return
}

func resourceShowHandler(c *fasthttp.RequestCtx, p fasthttprouter.Params) {
	c.SetContentType("application/json")

	if err := applyDelay(c, nil); err != nil {
		return
	}

	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		return
	}
	res, ok := resources[id]
	if !ok {
		c.SetStatusCode(http.StatusNotFound)
		fmt.Fprint(c, "Not Found")
		return
	}

	jsBytes, _ := json.Marshal(res)

	fmt.Fprint(c, string(jsBytes))

	return
}

func applyDelay(c *fasthttp.RequestCtx, _ fasthttprouter.Params) error {
	delay := string(c.Request.Header.Peek("X-Delay"))
	percentStr := string(c.Request.Header.Peek("X-Delay-Percent"))

	if delay == "" {
		return nil
	}

	duration, err := time.ParseDuration(delay)
	if err != nil {
		return err
	}

	percent := int64(100)
	percent, err = strconv.ParseInt(percentStr, 10, 0)
	if err != nil {
		percent = 100
	}

	r := rand.Int63n(100)

	if percent > r {
		time.Sleep(duration)
	}

	return nil
}

var (
	start = time.Now()
)

func applySlowdown(c *fasthttp.RequestCtx, _ fasthttprouter.Params) error {
	delayHeader := string(c.Request.Header.Peek("X-Slowdown"))
	fromTimeHeader := string(c.Request.Header.Peek("X-Slowdown-From"))

	if delayHeader == "" {
		return nil
	}
	if fromTimeHeader == "" {
		return nil
	}

	from, err := time.Parse(time.RFC3339, fromTimeHeader)
	if err != nil {
		return err
	}

	delay, err := time.ParseDuration(delayHeader)
	if err != nil {
		return err
	}

	if start.After(from) {
		time.Sleep(delay)
	}

	return nil
}

// Parse parses human readable bytes string to bytes integer.
// For example, 6GB (6G is also valid) will return 6442450944.
func sizeHandler(c *fasthttp.RequestCtx, p fasthttprouter.Params) {
	size, err := bytes.Parse(p.ByName("size"))
	if err != nil {
		return
	}

	if err := applyDelay(c, nil); err != nil {
		return
	}

	// not thread safe, so getting new src each time
	src := rand.NewSource(time.Now().UnixNano())
	fmt.Fprint(c, randStringBytesMaskImprSrc(int(size), src))
}

func fixedDelayResponse(c *fasthttp.RequestCtx, p fasthttprouter.Params) {
	duration, err := time.ParseDuration(p.ByName("delay"))
	if err != nil {
		// handle error
		return
	}

	time.Sleep(duration)

	return
}

func jsonHandler(c *fasthttp.RequestCtx, p fasthttprouter.Params) {

	if err := applyDelay(c, nil); err != nil {
		return
	}

	if err := applySlowdown(c, nil); err != nil {
		return
	}

	switch p.ByName("type") {
	case "invalid":
		fmt.Fprintf(c, `{time": "%s"}`, time.Now().String())
	default:
		fmt.Fprintf(c, `{"time": "%s"}`, time.Now().String())
	}

	return
}

func xmlHandler(c *fasthttp.RequestCtx, _ fasthttprouter.Params) {
	c.SetContentType("application/xml")

	fmt.Fprint(c, xml)

	return
}

func soapHandler(_ *fasthttp.RequestCtx, _ fasthttprouter.Params) {
	// TODO: SOAP Response

	return
}

func randStringBytesMaskImpr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func randStringBytesMaskImprSrc(n int, source rand.Source) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, source.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = source.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
