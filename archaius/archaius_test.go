// test archaius config
package archaius

import (
	"fmt"
	"testing"
	"time"
)

func TestConf(t *testing.T) {
	Conf.Arch = "testarch"
	Conf.GraphmlFile = "graphml"
	Conf.GraphjsonFile = "graphjson"
	Conf.RunDuration = 10 * time.Second
	Conf.Dunbar = 100
	Conf.Population = 100
	Conf.Msglog = true
	Conf.Regions = 2
	Conf.Collect = true
	Conf.StopStep = 2
	Conf.EurekaPoll = "1s"
	fmt.Println(string(AsJson()))
	FromJson(AsJson())
	fmt.Println(Conf)
}
