package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog"

	"flag"
	"fmt"
	"net/http"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/config"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/endpoints"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"
)

var cli struct {
	password      string
	configuration string
}

func init() {
	klog.InitFlags(nil)
	flag.StringVar(&cli.password, "pass", "examplePassword.yaml", "is the path to the password configuration")
	flag.StringVar(&cli.configuration, "config", "exampleConfiguration.yaml", "is the path to the configuration file")
}

func main() {
	flag.Parse()

	// config variables
	var pas models.Password
	var conf models.Configurations

	if err := config.ParseConfiguration(cli.configuration, &conf); err != nil {
		panic(err)
	}
	if err := config.ParseConfiguration(cli.password, &pas); err != nil {
		panic(err)
	}

	klog.Infof("configuration is parsed")

	klog.Infof("connect to database")
	var db database.Postgres
	if err := db.Connect(conf.Database.Address, pas.Database.User, pas.Database.Password, conf.Database.Database, conf.Database.Port); err != nil {
		panic(err)
	}

	klog.Infof("define server")
	var auth http.Handler
	auth = endpoints.Auth{Db: db}

	var contract http.Handler
	contract = endpoints.Contract{Db: db}

	// paths
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/analyses/", new(endpoints.Analyses))
	http.Handle("/machine-data/", new(endpoints.MachineData))
	http.Handle("/auth", auth)
	http.Handle("/health", new(endpoints.Health))
	http.Handle("/model/", new(endpoints.Model))
	http.Handle("/ready", new(endpoints.Ready))
	http.Handle("/contract/", contract)

	klog.Infof("start webserver")
	listen := fmt.Sprintf("%s:%d", conf.Webserver.Address, conf.Webserver.Port)
	_ = http.ListenAndServe(listen, nil)
}
