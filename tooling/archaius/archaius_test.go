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
	Conf.Neo4jURL = "localhost:7474"
	Conf.RunDuration = 10 * time.Second
	Conf.Dunbar = 100
	Conf.Population = 100
	Conf.Msglog = true
	Conf.Regions = 2
	Conf.Collect = true
	Conf.StopStep = 2
	Conf.EurekaPoll = "1s"
	Conf.Keyvals = "chat:0.01s"
	fmt.Println(string(AsJson()))
	FromJson(AsJson())
	fmt.Println(Conf)
	fmt.Println("chat = " + Key(Conf, "chat"))
}
