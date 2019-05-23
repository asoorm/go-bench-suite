package upstream

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/gommon/bytes"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
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

	return fasthttp.ListenAndServe(addr, r.HandleRequest)
}

func ServeTLS(addr, certFile, keyFile string) error {
	logrus.Infof("starting TLS server on %s", addr)

	r := registerRoutes()

	return fasthttp.ListenAndServeTLS(addr, certFile, keyFile, r.HandleRequest)
}

func registerRoutes() *routing.Router {
	r := routing.New()
	r.Get("/delay/<delay>", fixedDelayResponse)
	r.Get("/json/<type>", jsonHandler)
	r.Get("/xml", xmlHandler)
	r.Post("/soap", soapHandler)
	r.Any("/size/<size>", sizeHandler)

	seedResources()
	orchestrate := r.Group("/resource")
	orchestrate.Get("", resourceIndexHandler)
	orchestrate.Get("/", resourceIndexHandler)
	orchestrate.Get("/<id>", resourceShowHandler)

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

func resourceIndexHandler(c *routing.Context) error {
	c.SetContentType("application/json")

	if err := applyDelay(c); err != nil {
		return err
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

	return nil
}

func resourceShowHandler(c *routing.Context) error {
	c.SetContentType("application/json")

	if err := applyDelay(c); err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	res, ok := resources[id]
	if !ok {
		c.SetStatusCode(http.StatusNotFound)
		fmt.Fprint(c, "Not Found")
		return nil
	}

	jsBytes, _ := json.Marshal(res)

	fmt.Fprint(c, string(jsBytes))

	return nil
}

func applyDelay(c *routing.Context) error {
	delay := string(c.RequestCtx.Request.Header.Peek("X-Delay"))
	if delay != "" {
		duration, err := time.ParseDuration(delay)
		if err != nil {
			return err
		}
		time.Sleep(duration)
	}

	return nil
}

// Parse parses human readable bytes string to bytes integer.
// For example, 6GB (6G is also valid) will return 6442450944.
func sizeHandler(c *routing.Context) error {
	size, err := bytes.Parse(c.Param("size"))
	if err != nil {
		return err
	}

	if err := applyDelay(c); err != nil {
		return err
	}

	// not thread safe, so getting new src each time
	src := rand.NewSource(time.Now().UnixNano())
	fmt.Fprint(c, randStringBytesMaskImprSrc(int(size), src))

	return nil
}

func fixedDelayResponse(c *routing.Context) error {
	duration, err := time.ParseDuration(c.Param("delay"))
	if err != nil {
		return err
	}

	time.Sleep(duration)

	return nil
}

func jsonHandler(c *routing.Context) error {
	switch c.Param("type") {
	case "invalid":
		fmt.Fprintf(c, `{time": "%s"}`, time.Now().String())
	default:
		fmt.Fprintf(c, `{"time": "%s"}`, time.Now().String())
	}

	return nil
}

func xmlHandler(c *routing.Context) error {
	c.SetContentType("application/xml")

	fmt.Fprint(c, xml)

	return nil
}

func soapHandler(c *routing.Context) error {
	// TODO: SOAP Response

	return nil
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
