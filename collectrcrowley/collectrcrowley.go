// Collect throughput and response times using rcrowley Metrics
// Adrian - this package is really complex and hard to understand
// OK for doing exactly what it already does, but I couldn't figure out
// how to extend it to do what I need for spigo, so it isn't part of the build
// left here in case anyone else wants to play with it
package collectrcrowley

import (
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"github.com/rcrowley/go-metrics"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func NewHist(name string) metrics.Histogram {
	if name != "" && archaius.Conf.Collect {
		s := metrics.NewExpDecaySample(1028, 0.015)
		h := metrics.NewHistogram(s)
		metrics.Register(name, h)
		return h
	}
	return nil
}

func Measure(h metrics.Histogram, d time.Duration) {
	if h != nil && archaius.Conf.Collect {
		h.Update(int64(d))
	}
}

func Save() {
	if archaius.Conf.Collect {
		file, _ := os.Create(archaius.Conf.Arch + "_metrics.json")
		j, e := metrics.MarshalJSON()
		file.WriteString(string(j))
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
