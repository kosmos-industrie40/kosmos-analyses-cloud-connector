package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/config"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/auth"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/health"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/machineData"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/ready"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/mqtt"
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
	var pas config.Password
	var conf config.Configurations

	if err := config.ParseConfigurations(cli.configuration, &conf); err != nil {
		panic(err)
	}
	if err := config.ParsePassword(cli.password, &pas); err != nil {
		panic(err)
	}

	klog.Infof("configuration is parsed")

	klog.Infof("connect to database")
	conStr := fmt.Sprintf("host=%s user=%s password=%s port=%d sslmode=disable dbname=%s",
		conf.Database.Address,
		pas.Database.User,
		pas.Database.Password,
		conf.Database.Port,
		conf.Database.Database,
	)

	klog.Infof("conString: %s\n", conStr)
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		klog.Errorf("cannot connect to db: %s", err)
		os.Exit(1)
	}

	klog.Infof("connect to database")
	var mqttCo mqtt.Mqtt
	mqttCon := &mqttCo
	sendChan := make(chan mqtt.Msg, 100)
	er := make(chan error)
	if err := mqttCon.Init(pas.Mqtt.User, pas.Mqtt.Password, conf.Mqtt.Address, conf.Mqtt.Port, false, sendChan, er); err != nil {
		klog.Errorf("cannot connect to mqtt broker: %s", err)
		os.Exit(1)
	}

	go func() {
		for {
			e := <-er
			klog.Errorf("%v", e)
		}
	}()

	klog.Infof("setting up logic")
	/*
		var authentication logic.Authentication = logic.Auth{Db: db}
		var ana logic.Analyses = logic.AnalysesInitial{Db: db}
		var modelLogic logic.Model = logic.Mod{Db: db}
		var cont logic.Contract = logic.Contra{Db: db}
	*/
	//authentication.Authentication(db)
	//ana.Analyses(db)
	//modelLogic.Model(db)
	//cont.Contract(db)

	var authHandler http.Handler
	authHelper := auth.NewAuthHelper(db)
	authHandler, err  = auth.NewOidcAuth(conf.UserMgmt.UserMgmt, conf.UserMgmt.UserRealm, conf.UserMgmt.BasePath, pas.UserMgmt.ClientSecret, pas.UserMgmt.ClientId, conf.UserMgmt.ServerAddress)

	klog.Infof("define endpoints")
	machineHandler := machineData.NewMachineDataEndpoint(sendChan, authHelper)
	/*
		var auth http.Handler = machineData2.Auth{Auth: authentication}
		var contract http.Handler = contract2.Contract{Auth: authentication, Contract: cont}
		var machineData http.Handler = machineData2.MachineData{SendChan: sendChan, Auth: authentication}
		var analysesResult http.Handler = analysis.Analyses{Auth: authentication, Analyses: ana}
		var model http.Handler = model2.Model{Auth: authentication, Model: modelLogic}
	*/
	// paths
	http.Handle("/auth", authHandler)
	http.Handle("/health", new(health.Health))
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/machine-data", machineHandler)
	http.Handle("/ready", new(ready.Ready))

	//http.Handle("/analyses/", analysesResult)
	//http.Handle("/model/", model)
	//http.Handle("/contract/", contract)

	klog.Infof("start webserver")
	listen := fmt.Sprintf("%s:%d", conf.Webserver.Address, conf.Webserver.Port)
	_ = http.ListenAndServe(listen, nil)
}
