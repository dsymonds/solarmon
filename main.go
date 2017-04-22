package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	enphaseAddr = flag.String("enphase_addr", "", "IP address of Enphase Envoy")
	port        = flag.Int("port", 0, "port to run on")
)

// enphaseCollector implements prometheus.Collector.
type enphaseCollector struct {
	addr string

	up, genNow, genToday prometheus.Gauge
}

func newEnphaseCollector(addr string) *enphaseCollector {
	ec := &enphaseCollector{
		addr: addr,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "solar",
			Name:      "up",
			Help:      "Whether the Enphase Envoy is responding",
		}),
		genNow: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "solar",
			Name:      "power_production",
			Help:      "Power being produced, in W",
		}),
		genToday: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "solar",
			Name:      "energy_today",
			Help:      "Energy produced today, in Wh",
		}),
	}
	return ec
}

func (ec *enphaseCollector) Describe(ch chan<- *prometheus.Desc) {
	ec.up.Describe(ch)
	ec.genNow.Describe(ch)
	ec.genToday.Describe(ch)
}

func (ec *enphaseCollector) Collect(ch chan<- prometheus.Metric) {
	// TODO: rate limit here
	// TODO: set a HTTP timeout too. 5s should be plenty.
	stats, err := fetchStats(ec.addr)
	if err != nil {
		ec.up.Set(0)
		ec.genNow.Set(0)
		ec.genToday.Set(0)
	} else {
		ec.up.Set(1)
		ec.genNow.Set(float64(stats.WattsNow))
		ec.genToday.Set(float64(stats.WattHoursToday))
		log.Printf("Currently generating %.2f kW", float64(stats.WattsNow)/1e3)
		log.Printf("Generated %.1f kWh today", float64(stats.WattHoursToday)/1e3)
	}
	ch <- ec.up
	ch <- ec.genNow
	ch <- ec.genToday
}

func main() {
	flag.Parse()
	if *enphaseAddr == "" {
		log.Fatal("must provide -enphase_addr")
	}

	coll := newEnphaseCollector(*enphaseAddr)
	prometheus.MustRegister(coll)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
