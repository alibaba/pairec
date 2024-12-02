package prometheus

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/service/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

const (
	_          = iota // ignore first value by assigning to blank identifier
	KB float64 = 1 << (10 * iota)
	MB
)

var tr = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   100 * time.Millisecond, // 100ms
		KeepAlive: 5 * time.Minute,
	}).DialContext,
	MaxIdleConnsPerHost:   200,
	MaxIdleConns:          200,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 10 * time.Second,
}

var client = &http.Client{Transport: tr}

const defaultSubsystem = "pairec"

const defaultJob = "recommend"

// defaultReqDurBuckets is the buckets for request duration. Here, we use the prometheus defaults
// which are for ~10s request length max: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
var defaultReqDurBuckets = prometheus.DefBuckets

// defaultReqSzBuckets is the buckets for request size. Here we define a spectrom from 1KB thru 1NB up to 10MB.
var defaultReqSzBuckets = []float64{1.0 * KB, 2.0 * KB, 5.0 * KB, 10.0 * KB, 100 * KB, 500 * KB, 1.0 * MB, 2.5 * MB, 5.0 * MB, 10.0 * MB}

// defaultResSzBuckets is the buckets for response size. Here we define a spectrom from 1KB thru 1NB up to 10MB.
var defaultResSzBuckets = []float64{1.0 * KB, 2.0 * KB, 5.0 * KB, 10.0 * KB, 100 * KB, 500 * KB, 1.0 * MB, 2.5 * MB, 5.0 * MB, 10.0 * MB}

type RequestCounterLabelMappingFunc func(req *http.Request) string

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	reqCnt               *prometheus.CounterVec
	reqDur, reqSz, resSz *prometheus.HistogramVec

	ReqDurBuckets, ReqSzBuckets, ResSzBuckets []float64

	Subsystem string

	Ppg PushGateway
}

// PushGateway contains the configuration for pushing to a Prometheus pushgateway (optional)
type PushGateway struct {
	// Push interval in seconds
	//lint:ignore ST1011 renaming would be breaking change
	PushIntervalSeconds time.Duration

	// Push Gateway URL in format http://domain:port
	// where JOBNAME can be any string of your choice
	PushGatewayURL string

	// pushgateway job name, defaults to "recommend"
	Job string
}

// NewPrometheus generates a new set of metrics with a certain subsystem name
func NewPrometheus(options ...PrometheusOption) *Prometheus {
	p := &Prometheus{
		Subsystem: defaultSubsystem,

		ReqDurBuckets: defaultReqDurBuckets,
		ReqSzBuckets:  defaultReqSzBuckets,
		ResSzBuckets:  defaultResSzBuckets,
	}

	for _, option := range options {
		option(p)
	}

	p.registerMetrics()

	if p.Ppg.PushGatewayURL != "" && p.Ppg.PushIntervalSeconds > 0 {
		p.startPushTicker()
	}

	return p
}

func (p *Prometheus) Push(pushGatewayURL string, pushIntervalSecs int, job string) {
	p.Ppg.PushGatewayURL = pushGatewayURL
	p.Ppg.PushIntervalSeconds = time.Duration(pushIntervalSecs)

	if job != "" {
		p.Ppg.Job = job
	} else {
		p.Ppg.Job = defaultJob
	}

	if p.Ppg.PushGatewayURL != "" && p.Ppg.PushIntervalSeconds > 0 {
		p.startPushTicker()
	}
}

func (p *Prometheus) getMetrics() []byte {
	out := &bytes.Buffer{}
	metricFamilies, _ := prometheus.DefaultGatherer.Gather()
	for i := range metricFamilies {
		expfmt.MetricFamilyToText(out, metricFamilies[i])
	}

	return out.Bytes()
}

func (p *Prometheus) getCustomMetrics() []byte {
	out := &bytes.Buffer{}
	metricFamilies, _ := metrics.CustomRegister.Gather()
	for i := range metricFamilies {
		expfmt.MetricFamilyToText(out, metricFamilies[i])
	}

	return out.Bytes()
}

func (p *Prometheus) getPushGatewayURL() string {
	h, _ := os.Hostname()
	if p.Ppg.Job == "" {
		p.Ppg.Job = "recommend"
	}
	return p.Ppg.PushGatewayURL + "/metrics/job/" + p.Ppg.Job + "/instance/" + h
}

func (p *Prometheus) sendMetricsToPushGateway(metrics []byte) {
	req, err := http.NewRequest("POST", p.getPushGatewayURL(), bytes.NewBuffer(metrics))
	if err != nil {
		//log.Errorf("failed to create push gateway request: %v", err)
		return
	}

	if resp, err := client.Do(req); err != nil {
		log.Error(fmt.Sprintf("Error sending to push gateway: %v", err))
	} else {
		resp.Body.Close()
	}
}

func (p *Prometheus) startPushTicker() {
	ticker := time.NewTicker(time.Second * p.Ppg.PushIntervalSeconds)
	go func() {
		for range ticker.C {
			p.sendMetricsToPushGateway(p.getMetrics())

			p.sendMetricsToPushGateway(p.getCustomMetrics())
		}
	}()
}

func (p *Prometheus) registerMetrics() {
	subsystem := p.Subsystem

	p.reqCnt = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"code", "method", "host", "url"},
	)

	p.reqDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "The HTTP request latencies in seconds.",
			Buckets:   p.ReqDurBuckets,
		},
		[]string{"code", "method", "host", "url"},
	)

	p.reqSz = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "request_size_bytes",
			Help:      "The HTTP request sizes in bytes.",
			Buckets:   p.ReqSzBuckets,
		},
		[]string{"code", "method", "host", "url"},
	)

	p.resSz = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "response_size_bytes",
			Help:      "The HTTP response sizes in bytes.",
			Buckets:   p.ResSzBuckets,
		},
		[]string{"code", "method", "host", "url"},
	)

	collectors := map[string]prometheus.Collector{
		"requests_total":           p.reqCnt,
		"request_duration_seconds": p.reqDur,
		"request_size_bytes":       p.reqSz,
		"response_size_bytes":      p.resSz,
	}

	for name := range collectors {
		if err := prometheus.Register(collectors[name]); err != nil {
			log.Error(fmt.Sprintf("%s could not be registered in Prometheus: %v", name, err))
		}
	}
}

type IResponseWriter interface {
	http.ResponseWriter

	StatusCode() int
	Size() int
}

// HandlerFunc defines handler function for middleware
func (p *Prometheus) HandlerFunc(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		reqSz := computeApproximateRequestSize(request)

		next(writer, request)

		status := http.StatusOK
		var size int
		if rw, ok := writer.(IResponseWriter); ok {
			status = rw.StatusCode()
			size = rw.Size()
		}

		elapsed := float64(time.Since(start)) / float64(time.Second)

		url := request.URL.Path
		host := request.Host

		statusStr := strconv.Itoa(status)

		p.reqDur.WithLabelValues(statusStr, request.Method, host, url).Observe(elapsed)
		p.reqCnt.WithLabelValues(statusStr, request.Method, host, url).Inc()
		p.reqSz.WithLabelValues(statusStr, request.Method, host, url).Observe(float64(reqSz))

		resSz := float64(size)
		p.resSz.WithLabelValues(statusStr, request.Method, host, url).Observe(resSz)
	}
}

func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
