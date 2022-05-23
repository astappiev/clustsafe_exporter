package main

import (
	"encoding/xml"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"fmt"
	"os/exec"
)

type dataCollector struct {
	clustsafeHost, clustsafeUser, clustsafePassword string
	humidity, temperature *prometheus.Desc
}

func ClustsafeCollector(clustsafeHost string, clustsafeUser string, clustsafePassword string) *dataCollector {
	return &dataCollector{
		clustsafeHost:     clustsafeHost,
		clustsafeUser:     clustsafeUser,
		clustsafePassword: clustsafePassword,

		humidity:     prometheus.NewDesc("clustsafe_humidity", "The humidity in percentage", []string{"sensor"}, nil),
		temperature:  prometheus.NewDesc("clustsafe_temperature", "The temperature in celsius", []string{"sensor"}, nil),
	}
}

func (e *dataCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.humidity
	ch <- e.temperature
}

func (e *dataCollector) Collect(ch chan<- prometheus.Metric) {
	out, err := exec.Command("cw-clustsafe",
		"--host", e.clustsafeHost,
		"--user", e.clustsafeUser,
		"--password", e.clustsafePassword,
		"-a", "-x", "sensors").Output()

	if err != nil {
		log.Fatal(err)
	}

	var response ClustsafeResponse
	xml.Unmarshal(out, &response)

	for i := 0; i < len(response.Sensors); i++ {
		if response.Sensors[i].Type == "humidity" {
			ch <- prometheus.MustNewConstMetric(e.humidity, prometheus.GaugeValue, float64(response.Sensors[i].Value), "humidity")
		} else if response.Sensors[i].Type == "temperature" {
			ch <- prometheus.MustNewConstMetric(e.temperature, prometheus.GaugeValue, float64(response.Sensors[i].Value), "temperature")
		} else if response.Sensors[i].Type == "dallas" && response.Sensors[i].Id > 0 {
			ch <- prometheus.MustNewConstMetric(e.temperature, prometheus.GaugeValue, float64(response.Sensors[i].Value), fmt.Sprint("specific", response.Sensors[i].Id))
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
