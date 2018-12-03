package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	mp "github.com/mackerelio/go-mackerel-plugin"
)

// MongoDBMongoDBSlowQueriesPluginPlugin mackerel plugin for mongo
type MongoDBSlowQueriesPlugin struct {
	Prefix   string
	URL      string
	Username string
	Password string
	Database string
}

func (m MongoDBSlowQueriesPlugin) MetricKeyPrefix() string {
	if m.Prefix == "" {
		m.Prefix = "mongodb"
	}
	return m.Prefix
}

// GraphDefinition interface for mackerelplugin
func (m MongoDBSlowQueriesPlugin) GraphDefinition() map[string]mp.Graphs {
	return map[string]mp.Graphs{
		"slow_queries": {
			Label: "MongoDB Slow Queries",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "count", Label: "Slow Queries"},
			},
		},
	}
}

// FetchMetrics interface for mackerelplugin
func (m MongoDBSlowQueriesPlugin) FetchMetrics() (map[string]float64, error) {
	session, err := mgo.Dial(m.URL)
	if err != nil {
		return nil, err
	}

	collection := session.DB(m.Database).C("system.profile")
	one_minute_ago := time.Unix(time.Now().Unix()-60, 0)

	count, err := collection.Find(bson.M{"ts": bson.M{"$gt": one_minute_ago}}).Count()
	if err != nil {
		return nil, err
	}

	return map[string]float64{"count": float64(count)}, err
}

// Do the plugin
func main() {
	optPrefix := flag.String("metric-key-prefix", "mongodb", "Metric key prefix")
	optHost := flag.String("host", "localhost", "Hostname")
	optPort := flag.String("port", "27017", "Port")
	optUser := flag.String("username", "", "Username")
	optPass := flag.String("password", os.Getenv("MONGODB_PASSWORD"), "Password")
	optDatabase := flag.String("database", "", "Database name")
	flag.Parse()

	var mongodb MongoDBSlowQueriesPlugin
	mongodb.Prefix = *optPrefix
	mongodb.URL = fmt.Sprintf("%s:%s", *optHost, *optPort)
	mongodb.Username = fmt.Sprintf("%s", *optUser)
	mongodb.Password = fmt.Sprintf("%s", *optPass)
	mongodb.Database = fmt.Sprintf("%s", *optDatabase)

	helper := mp.NewMackerelPlugin(mongodb)

	helper.Run()
}
