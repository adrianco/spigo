// graphneo4j tests
package graphneo4j

import (
	"github.com/adrianco/spigo/tooling/archaius"
	"github.com/adrianco/spigo/tooling/names"
	"testing"
	"time"
)

// reader parses graphjson
func TestGraph(t *testing.T) {
	testNeo := `
	  CREATE (test_mysql00:test:store:mysql {name:"mysql00", node:"test.us-east-1.zoneA..mysql00...mysql.store", timestamp:"2016-04-17T13:40:05.938437713-07:00", ip:"54.198.0.1", region:"us-east-1", zone: "zoneA"}),
          (test_mysql01:test:store:mysql {name:"mysql01", node:"test.us-east-1.zoneA..mysql01...mysql.store", timestamp:"2016-04-17T13:40:05.938513762-07:00", ip:"54.221.0.1", region:"us-east-1", zone: "zoneA"}),
          (test_mysql00)-[:CONN]->(test_mysql01)
          `
	archaius.Conf.Arch = "test"
	Setup("localhost:7474")
	Write(testNeo)
	dal0 := names.Make("test", "us-east-1", "ZoneA", "dal", "staash", 0)
	WriteNode(dal0+" staash", time.Now())
	WriteEdge(dal0+" test.us-east-1.zoneA..mysql00...mysql.store", time.Now())
	WriteEdge(dal0+" test.us-east-1.zoneA..mysql01...mysql.store", time.Now())
	WriteFlow(dal0, "test.us-east-1.zoneA..mysql00...mysql.store", "Put", 100, 1)
	Close()
}
