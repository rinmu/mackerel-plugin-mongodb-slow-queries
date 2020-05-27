package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/globalsign/mgo"

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
		"slow_queries_total": {
			Label: "MongoDB Slow Queries",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "total_time", Label: "Slow Queries Total Time"},
			},
		},
		"slow_queries_average": {
			Label: "MongoDB Slow Queries",
			Unit:  "float",
			Metrics: []mp.Metrics{
				{Name: "average_time", Label: "Slow Queries Average Time"},
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

	session.SetMode(mgo.Nearest, true)

	one_minute_ago := time.Now().Add(time.Duration(-1) * time.Minute).Format(time.RFC3339)

	command := fmt.Sprintf(`mongo %s --eval `, m.Database)
	query := fmt.Sprintf(`'rs.slaveOk(); db.system.profile.aggregate([
		{$match: {ts: {$gt: ISODate("%s")}}},
		{$group: {_id: "", average_time: {$avg: "$millis"}, count: {$sum:1}, total_time: {$sum:"$millis"}}}
	]);'`, one_minute_ago)
	grep := "|grep average_time"

	out, err := exec.Command("bash", "-c", command+query+grep).Output()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	re := regexp.MustCompile(`"average_time" : (.*?),`)
	match := re.FindStringSubmatch(string(out))
	averageTime, err := strconv.ParseFloat(match[1], 64)

	re = regexp.MustCompile(`"count" : (.*?),`)
	match = re.FindStringSubmatch(string(out))
	count, err := strconv.ParseFloat(match[1], 64)

	re = regexp.MustCompile(`"total_time" : (.*?)\s}`)
	match = re.FindStringSubmatch(string(out))
	totalTime, err := strconv.ParseFloat(match[1], 64)

	if err != nil {
		return nil, err
	}

	return map[string]float64{
		"count":        count,
		"total_time":   totalTime,
		"average_time": averageTime,
	}, err
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
