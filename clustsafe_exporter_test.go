package main

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/astappiev/clustsafe_exporter/collector"
)

func testConfig(t *testing.T) *collector.Config {
	return &collector.Config{
		FetchFn: func(file string) ([]byte, error) {
			exp, err := os.Open(path.Join("testdata", file))
			if err != nil {
				t.Fatalf("Error opening fixture file %q: %v", file, err)
			}

			return io.ReadAll(exp)
		},
	}
}

func expectMetrics(t *testing.T, c prometheus.Collector, fixture string) {
	exp, err := os.Open(path.Join("testdata", fixture))
	if err != nil {
		t.Fatalf("Error opening fixture file %q: %v", fixture, err)
	}
	if err := testutil.CollectAndCompare(c, exp); err != nil {
		t.Fatal("Unexpected metrics returned:", err)
	}
}

func TestAll(t *testing.T) {
	e := collector.NewExporter("all.xml", testConfig(t), log.NewNopLogger())
	expectMetrics(t, e, "all.metrics")
}

func TestClustsafes(t *testing.T) {
	e := collector.NewExporter("clustsafes.xml", testConfig(t), log.NewNopLogger())
	expectMetrics(t, e, "clustsafes.metrics")
}

func TestSensors(t *testing.T) {
	e := collector.NewExporter("sensors.xml", testConfig(t), log.NewNopLogger())
	expectMetrics(t, e, "sensors.metrics")
}
