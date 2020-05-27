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
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/mqtt"
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

	klog.Infof("connect to database")
	var mqttCo mqtt.Mqtt
	mqttCon := &mqttCo
	sendChan := make(chan mqtt.Msg, 100)
	er := make(chan error)
	if err := mqttCon.Init(pas.Mqtt.User, pas.Mqtt.Password, conf.Mqtt.Address, conf.Mqtt.Port, false, sendChan, er); err != nil {
		panic(err)
	}

	go func() {
		for {
			e := <-er
			klog.Errorf("%v", e)
		}
	}()

	klog.Infof("define server")
	var auth http.Handler = endpoints.Auth{Db: db}

	var contract http.Handler = endpoints.Contract{Db: db}

	var machineData http.Handler = endpoints.MachineData{SendChan: sendChan, Db: db}

	var analysesResult http.Handler = endpoints.Analyses{Db: db}

	// paths
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/analyses/", analysesResult)
	http.Handle("/machine-data/", machineData)
	http.Handle("/auth", auth)
	http.Handle("/health", new(endpoints.Health))
	http.Handle("/model/", new(endpoints.Model))
	http.Handle("/ready", new(endpoints.Ready))
	http.Handle("/contract/", contract)

	klog.Infof("start webserver")
	listen := fmt.Sprintf("%s:%d", conf.Webserver.Address, conf.Webserver.Port)
	_ = http.ListenAndServe(listen, nil)
}
