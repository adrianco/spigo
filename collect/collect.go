// Collect throughput and response times using Metrics
package collect

import (
	"encoding/json"
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"github.com/codahale/metrics"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func NewHist(name string) *metrics.Histogram {
	if name != "" && archaius.Conf.Collect {
		return metrics.NewHistogram(name, 1000, 100000000, 5)
	}
	return nil
}

func Measure(h *metrics.Histogram, d time.Duration) {
	if h != nil && archaius.Conf.Collect {
		h.RecordValue(int64(d))
		metrics.Counter(h.Name()).Add()
	}
}

func Save() {
	if archaius.Conf.Collect {
		file, _ := os.Create(archaius.Conf.Arch + "_metrics.json")
		counters, gauges := metrics.Snapshot()
		cj, _ := json.Marshal(counters)
		gj, _ := json.Marshal(gauges)
		file.WriteString(fmt.Sprintf("{\n\"counters\":%v\n\"gauges\":%v\n}\n", string(cj), string(gj)))
		file.Close()
	}
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
