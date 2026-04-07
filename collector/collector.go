package collector

import (
	"encoding/xml"
	"log/slog"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	Path     string
	Command  string
	User     string
	Password string
	FetchFn  func(string) ([]byte, error)
}

type Exporter struct {
	clustsafeHost string
	fetchFn       func(string) ([]byte, error)

	humidity         *prometheus.Desc
	temperature      *prometheus.Desc
	outletStatus     *prometheus.Desc
	lineStatus       *prometheus.Desc
	powerConsumption *prometheus.Desc
	status           *prometheus.Desc
	logger           *slog.Logger
}

func NewExporter(clustsafeHost string, config *Config, logger *slog.Logger) *Exporter {
	if config.FetchFn == nil {
		config.FetchFn = func(host string) ([]byte, error) {
			return exec.Command(config.Path, "--host", host,
				"--user", config.User, "--password", config.Password,
				"-a",           // accept all SSL certificates
				"-x",           // output in XML format
				config.Command, // request information (default: sensors)
			).Output()
		}
	}

	return &Exporter{
		clustsafeHost: clustsafeHost,
		fetchFn:       config.FetchFn,
		logger:        logger,

		humidity:         prometheus.NewDesc("clustsafe_humidity", "The humidity in percentage.", []string{"sensor"}, nil),
		temperature:      prometheus.NewDesc("clustsafe_temperature", "The temperature in celsius.", []string{"sensor"}, nil),
		outletStatus:     prometheus.NewDesc("clustsafe_outlet_up", "The status of an outlet in the module.", []string{"module", "outlet"}, nil),
		lineStatus:       prometheus.NewDesc("clustsafe_line_up", "The status of in input power line in the module.", []string{"module", "line"}, nil),
		powerConsumption: prometheus.NewDesc("clustsafe_power_consumption_watts", "The real power consumption in Watts.", []string{"module"}, nil),
		status:           prometheus.NewDesc("clustsafe_up", "Was the last scrape of ClustSafe successful.", nil, nil),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.humidity
	ch <- e.temperature
	ch <- e.outletStatus
	ch <- e.lineStatus
	ch <- e.powerConsumption
	ch <- e.status
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) (up float64) {
	out, err := e.fetchFn(e.clustsafeHost)

	if err != nil {
		e.logger.Error("Can't execute clustsafe command", "err", err)
		return 0
	}

	var data ClustsafeSchema
	if err := xml.Unmarshal(out, &data); err != nil {
		e.logger.Debug("Received clustsafe output", "output", string(out))
		e.logger.Error("Failed parsing clustsafe output", "err", err)
		return 0
	}

	for _, module := range data.Modules.Clustsafe {
		if module.Status != "connected" {
			continue
		}

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

	for _, sensor := range data.Sensors.Sensor {
		switch sensor.Type {
		case "humidity":
			ch <- prometheus.MustNewConstMetric(e.humidity, prometheus.GaugeValue, sensor.Value, "humidity")
		case "temperature":
			ch <- prometheus.MustNewConstMetric(e.temperature, prometheus.GaugeValue, sensor.Value, "temperature")
		case "dallas":
			if sensor.Value > 0 {
				ch <- prometheus.MustNewConstMetric(e.temperature, prometheus.GaugeValue, sensor.Value, "dallas"+sensor.ID)
			}
		}
	}

	return 1
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	up := e.scrape(ch)

	ch <- prometheus.MustNewConstMetric(e.status, prometheus.GaugeValue, up)
}
