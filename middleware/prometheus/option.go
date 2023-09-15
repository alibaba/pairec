package prometheus

type PrometheusOption func(p *Prometheus)

var WithSubsystem = func(subsystem string) PrometheusOption {
	return func(p *Prometheus) {
		if subsystem != "" {
			p.Subsystem = subsystem
		}
	}
}

var WithReqDurBuckets = func(buckets []float64) PrometheusOption {
	return func(p *Prometheus) {
		if len(buckets) > 0 {
			p.ReqDurBuckets = buckets
		}
	}
}

var WithReqSzBuckets = func(buckets []float64) PrometheusOption {
	return func(p *Prometheus) {
		if len(buckets) > 0 {
			p.ReqSzBuckets = buckets
		}
	}
}

var WithResSzBuckets = func(buckets []float64) PrometheusOption {
	return func(p *Prometheus) {
		if len(buckets) > 0 {
			p.ResSzBuckets = buckets
		}
	}
}
