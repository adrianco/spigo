// compose tests - just make sure the yaml conversions work
package compose

import (
	//"encoding/json"
	"fmt"
	"github.com/adrianco/spigo/archaius" // global configuration
	//"github.com/adrianco/spigo/architecture"
	"gopkg.in/yaml.v2"
	"testing"
	"time"
)


func try(t string) {
	var c ComposeYaml
	err := yaml.Unmarshal([]byte(t), &c)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(*c)
	for i, v := range c {
		fmt.Println("Compose: ", i, v.Build, v.Links)
	}
	
}

// test based on https://github.com/b00giZm/docker-compose-nodejs-examples/blob/master/05-nginx-express-redis-nodemon/docker-compose.yml
func TestGraph(t *testing.T) {
	testyaml := `
web:
  build: ./app
  volumes:
    - "app:/src/app"
  expose:
    - "3000"
  links:
    - "db:redis"
  command: nodemon -L app/bin/www

nginx:
  restart: always
  build: ./nginx/
  ports:
    - "80:80"
  volumes:
    - /www/public
  volumes_from:
    - web
  links:
    - web:web

db:
  image: redis
`

	archaius.Conf.Arch = "test"
	//archaius.Conf.GraphmlFile = ""
	//archaius.Conf.GraphjsonFile = ""
	archaius.Conf.RunDuration = 2 * time.Second
	archaius.Conf.Dunbar = 50
	archaius.Conf.Population = 50
	//archaius.Conf.Msglog = false
	archaius.Conf.Regions = 1
	//archaius.Conf.Collect = false
	//archaius.Conf.StopStep = 0
	archaius.Conf.EurekaPoll = "1s"
	try(testyaml)
	ReadCompose("test")
	//fmt.Println(*a)
	//Start(a)
}
