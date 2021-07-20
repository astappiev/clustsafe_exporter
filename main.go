package main

import (
	"encoding/xml"
	"flag"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var (
	listenAddress = flag.String("web.listen-address", ":9879", "Address to listen on for telemetry")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
)

type Exporter struct {
	clustsafeHost, clustsafeUser, clustsafePassword string
	humidity, temperature, temperature1             *prometheus.Desc
}

func ClustsafeExporter(clustsafeHost string, clustsafeUser string, clustsafePassword string) *Exporter {
	return &Exporter{
		clustsafeHost:     clustsafeHost,
		clustsafeUser:     clustsafeUser,
		clustsafePassword: clustsafePassword,

		humidity:     prometheus.NewDesc("clustsafe_humidity", "Shows the current humidity in percentage, a normal value is up to 80", []string{"host"}, nil),
		temperature:  prometheus.NewDesc("clustsafe_temperature", "The current temperature in celsius, a normal value is up to 27", []string{"host"}, nil),
		temperature1: prometheus.NewDesc("clustsafe_temperature1", "The current temperature in celsius (on top of the rack), a normal value is up to 35", []string{"host"}, nil),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.humidity
	ch <- e.temperature
	ch <- e.temperature1
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	hosts := strings.Split(e.clustsafeHost, ",")

	for h := 0; h < len(hosts); h++ {
		out, err := exec.Command("clustsafeX", "--host", hosts[h],
			"--user", e.clustsafeUser, "--password", e.clustsafePassword, "-a", "-x", "sensors").Output()

		if err != nil {
			log.Fatal(err)
		}

		var response ClustsafeResponse
		xml.Unmarshal(out, &response)

		for i := 0; i < len(response.Sensors); i++ {
			if response.Sensors[i].Type == "humidity" {
				ch <- prometheus.MustNewConstMetric(e.humidity, prometheus.CounterValue, float64(response.Sensors[i].Value), hosts[h])
			} else if response.Sensors[i].Type == "temperature" {
				ch <- prometheus.MustNewConstMetric(e.temperature, prometheus.CounterValue, float64(response.Sensors[i].Value), hosts[h])
			} else if response.Sensors[i].Type == "dallas" && response.Sensors[i].Id == 1 {
				ch <- prometheus.MustNewConstMetric(e.temperature1, prometheus.CounterValue, float64(response.Sensors[i].Value), hosts[h])
			}
		}
	}
}

type ClustsafeResponse struct {
	XMLName xml.Name `xml:"clustsafeResponse"`
	Sensors []Sensor `xml:"sensors>sensor"`
}

type Sensor struct {
	XMLName xml.Name `xml:"sensor"`
	Id      int      `xml:"id,attr"`
	Type    string   `xml:"type,attr"`
	Status  string   `xml:"status"`
	Value   int      `xml:"value"`
}

func main() {
	flag.Parse()

	godotenv.Load()
	clustsafeHost := os.Getenv("CLUSTSAFE_HOST")
	clustsafeUser := os.Getenv("CLUSTSAFE_USER")
	clustsafePassword := os.Getenv("CLUSTSAFE_PASSWORD")

	if len(clustsafeHost) == 0 {
		log.Println("No host given! Check env variables.")
	}

	exporter := ClustsafeExporter(clustsafeHost, clustsafeUser, clustsafePassword)
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
