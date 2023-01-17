package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
)

type CollectorConfig struct {
	clustsafePath     string
	clustsafeCommand  string
	clustsafeUser     string
	clustsafePassword string
}

type dataCollector struct {
	host             string
	config           CollectorConfig
	humidity         *prometheus.Desc
	temperature      *prometheus.Desc
	outletStatus     *prometheus.Desc
	lineStatus       *prometheus.Desc
	powerConsumption *prometheus.Desc
}

func ClustsafeCollector(clustsafeHost string, config CollectorConfig) *dataCollector {
	return &dataCollector{
		host:             clustsafeHost,
		config:           config,
		humidity:         prometheus.NewDesc("clustsafe_humidity", "The humidity in percentage", []string{"sensor"}, nil),
		temperature:      prometheus.NewDesc("clustsafe_temperature", "The temperature in celsius", []string{"sensor"}, nil),
		outletStatus:     prometheus.NewDesc("clustsafe_outlet_up", "The status of an outlet in the module", []string{"module", "outlet"}, nil),
		lineStatus:       prometheus.NewDesc("clustsafe_line_up", "The status of in input power line in the module", []string{"module", "line"}, nil),
		powerConsumption: prometheus.NewDesc("clustsafe_power_consumption_watts", "The real power consumption in Watts", []string{"module"}, nil),
	}
}

func (e *dataCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.humidity
	ch <- e.temperature
	ch <- e.outletStatus
	ch <- e.lineStatus
	ch <- e.powerConsumption
}

func (e *dataCollector) Collect(ch chan<- prometheus.Metric) {
	out, err := exec.Command(e.config.clustsafePath, "--host", e.host,
		"--user", e.config.clustsafeUser, "--password", e.config.clustsafePassword,
		"-a",                      // accept all SSL certificates
		"-x",                      // output in XML format
		e.config.clustsafeCommand, // request information (default: sensors)
	).Output()

	if err != nil {
		err := fmt.Errorf("executing command error: %w", err)
		ch <- prometheus.NewInvalidMetric(prometheus.NewInvalidDesc(err), nil)
		log.Println(err)
		return
	}

	var response ClustsafeResponse
	err = xml.Unmarshal(out, &response)
	if err != nil {
		err := fmt.Errorf("parsing response error: %w\n\nresponse: %s", err, out)
		ch <- prometheus.NewInvalidMetric(prometheus.NewInvalidDesc(err), nil)
		log.Println(err)
		return
	}

	for _, module := range response.Modules.Clustsafe {
		if module.Status == "connected" {
			ch <- prometheus.MustNewConstMetric(e.powerConsumption, prometheus.GaugeValue, module.Power.RealPower, module.ID)

			for _, outlet := range module.Outlets.Outlet {
				var status float64
				if outlet.Status == "on" {
					status = 1.0
				}

				ch <- prometheus.MustNewConstMetric(e.outletStatus, prometheus.GaugeValue, status, module.ID, outlet.ID)
			}

			for _, line := range module.Lines.Line {
				var status float64
				if line.Status == "connected" {
					status = 1.0
				}

				ch <- prometheus.MustNewConstMetric(e.lineStatus, prometheus.GaugeValue, status, module.ID, line.ID)
			}
		}
	}

	for _, sensor := range response.Sensors.Sensor {
		if sensor.Type == "humidity" {
			ch <- prometheus.MustNewConstMetric(e.humidity, prometheus.GaugeValue, sensor.Value, "humidity")
		} else if sensor.Type == "temperature" {
			ch <- prometheus.MustNewConstMetric(e.temperature, prometheus.GaugeValue, sensor.Value, "temperature")
		} else if sensor.Type == "dallas" && sensor.Value > 0 {
			ch <- prometheus.MustNewConstMetric(e.temperature, prometheus.GaugeValue, sensor.Value, "dallas"+sensor.ID)
		}
	}
}

type ClustsafeResponse struct {
	Modules struct {
		Clustsafe []struct {
			ID      string `xml:"id,attr"`
			Status  string `xml:"status"`
			Power   Power  `xml:"power"`
			Outlets struct {
				Count  string `xml:"count,attr"`
				Outlet []struct {
					ID        string `xml:"id,attr"`
					Status    string `xml:"status"`
					Fuse      string `xml:"fuse"`
					Autopower string `xml:"autopower"`
					Power     Power  `xml:"power"`
				} `xml:"outlet"`
			} `xml:"outlets"`
			Lines struct {
				Count string `xml:"count,attr"`
				Line  []struct {
					ID             string `xml:"id,attr"`
					Status         string `xml:"status"`
					Identification string `xml:"identification"`
					Power          Power  `xml:"power"`
				} `xml:"line"`
			} `xml:"lines"`
		} `xml:"clustsafe"`
	} `xml:"modules"`
	Sensors struct {
		Count  string `xml:"count,attr"`
		Sensor []struct {
			Type       string  `xml:"type,attr"`
			ID         string  `xml:"id,attr"`
			Status     string  `xml:"status"`
			Value      float64 `xml:"value"`
			Alert      string  `xml:"alert"`
			Identifier string  `xml:"identifier"`
		} `xml:"sensor"`
	} `xml:"sensors"`
}

type Power struct {
	Status        string  `xml:"status,attr"`
	Voltage       float64 `xml:"voltage"`
	Current       float64 `xml:"current"`
	Frequency     float64 `xml:"frequency"`
	RealPower     float64 `xml:"realPower"`
	ApparentPower float64 `xml:"apparentPower"`
	PowerFactor   float64 `xml:"powerFactor"`
	PhaseShift    float64 `xml:"phaseShift"`
	Samples       float64 `xml:"samples"`
}
