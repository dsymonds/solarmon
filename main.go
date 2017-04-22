package main

import (
	"flag"
	"log"
)

var (
	enphaseAddr = flag.String("enphase_addr", "", "IP address of Enphase Envoy")
)

func main() {
	flag.Parse()
	if *enphaseAddr == "" {
		log.Fatal("must provide -enphase_addr")
	}

	stats, err := fetchStats(*enphaseAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Currently generating %.2f kW", float64(stats.WattsNow)/1e3)
	log.Printf("Generated %.1f kWh today", float64(stats.WattHoursToday)/1e3)
}
