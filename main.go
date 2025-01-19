package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	enphaseAddr = flag.String("enphase_addr", "", "IP address of Enphase Envoy")
	port        = flag.Int("port", 0, "port to run on")

	timezones = flag.String("timezones", "", "comma-separated list of timezones to export gauges for")
)

var (
	up = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "up",
		Help: "Whether the Enphase Envoy is responding",
	})
	genNow = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "power_production_watts",
		Help: "Power being produced, in W",
	})
	genToday = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "energy_today_watthours",
		Help: "Energy produced today, in Wh",
	})
	// In theory this should always increase, but we can't guarantee it,
	// so leave as a gauge instead of a counter. It's only read as snapshots anyway.
	genLifetime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "energy_lifetime_watthours",
		Help: "Energy produced in the system's lifetime, in Wh",
	})
	localTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "local_time",
		Help: "Local time (hour and fractional minute)",
	}, []string{"tz"})
)

// enphaseMonitor monitors an Enphase Envoy.
type enphaseMonitor struct {
	addr string
	tzs  map[string]*time.Location
}

func newEnphaseMonitor(addr string) (*enphaseMonitor, error) {
	tzs := make(map[string]*time.Location)
	if *timezones != "" {
		for _, tz := range strings.Split(*timezones, ",") {
			loc, err := time.LoadLocation(tz)
			if err != nil {
				return nil, fmt.Errorf("Loading location %q: %w", tz, err)
			}
			tzs[tz] = loc
		}
	}

	return &enphaseMonitor{
		addr: addr,
		tzs:  tzs,
	}, nil
}

func (em *enphaseMonitor) Refresh() {
	// TODO: rate limit here
	// TODO: set a HTTP timeout too. 5s should be plenty.
	stats, err := fetchStats(em.addr)
	if err != nil {
		up.Set(0)
		genNow.Set(0)
		genToday.Set(0)
		genLifetime.Set(0)
	} else {
		up.Set(1)
		genNow.Set(float64(stats.WattsNow))
		genToday.Set(float64(stats.WattHoursToday))
		genLifetime.Set(float64(stats.WattHoursLifetime))
		//log.Printf("Currently generating %.2f kW", float64(stats.WattsNow)/1e3)
		//log.Printf("Generated %.1f kWh today", float64(stats.WattHoursToday)/1e3)
	}

	now := time.Now()
	for tz, loc := range em.tzs {
		h, m, _ := now.In(loc).Clock()
		localTime.WithLabelValues(tz).Set(float64(h) + float64(m)/60)
	}
}

func main() {
	flag.Parse()
	if *enphaseAddr == "" {
		log.Fatal("must provide -enphase_addr")
	}

	mon, err := newEnphaseMonitor(*enphaseAddr)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		mon.Refresh()
		for range time.NewTicker(1 * time.Minute).C {
			mon.Refresh()
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
