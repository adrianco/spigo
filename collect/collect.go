// Collect throughput and response times using Go-Kit Metrics
package collect

import (
	//	"encoding/json"
	"fmt"
	"github.com/adrianco/kit/metrics/expvar"
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/names"
	"github.com/go-kit/kit/metrics"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	maxHistObservable = 1000000
)

//var mon = monitor.GetMonitors()

func NewHist(name string) metrics.Histogram {
	var h metrics.Histogram
	if name != "" && archaius.Conf.Collect {
		h = expvar.NewHistogram(name, 1000, maxHistObservable, 1, []int{50, 99}...)
		return h
	}
	return nil
}

func Measure(h metrics.Histogram, d time.Duration) {
	if h != nil && archaius.Conf.Collect {
		if d > maxHistObservable {
			h.Observe(int64(maxHistObservable))
		} else {
			h.Observe(int64(d))
		}
	}
}

// have to pass in name because metrics.Histogram blocks expvar.Historgram.Name()
func SaveHist(h metrics.Histogram, name, suffix string) {
	if archaius.Conf.Collect {
		file, err := os.Create("csv_metrics/" + names.Arch(name) + "_" + names.Machine(name) + suffix + ".csv")
		if err != nil {
			log.Printf("%v: %v\n", name, err)
		}
		file.WriteString(fmt.Sprintf("%v", h))
		file.Close()
	}
}

func Save() {
	//	if archaius.Conf.Collect {
	//		file, _ := os.Create("csv_metrics/" + archaius.Conf.Arch + "_metrics.csv")
	//		counters, gauges := metrics.Snapshot()
	//		cj, _ := json.Marshal(counters)
	//		gj, _ := json.Marshal(gauges)
	//		file.WriteString(fmt.Sprintf("{\n\"counters\":%v\n\"gauges\":%v\n}\n", string(cj), string(gj)))
	//		file.Close()
	//	}
}

func Serve(port int) {
	sock, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		log.Printf("HTTP metrics now available at localhost:%v/debug/vars", port)
		http.Serve(sock, nil)
	}()
}
