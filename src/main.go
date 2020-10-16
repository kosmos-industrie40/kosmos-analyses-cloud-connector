package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/config"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/analysis"
	analysisModel "gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/analysis/models"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/auth"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/contract"
	contractModel "gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/contract/models"
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

func dbVersion(db *sql.DB) {
	version, err := db.Query("SELECT version()")
	if err != nil {
		klog.Errorf("cannot select db version:  %s", err)
		os.Exit(1)
	}

	defer func() {
		if err := version.Close(); err != nil {
			klog.Errorf("cannot close db query: %s", err)
			os.Exit(1)
		}
	}()

	if !version.Next() {
		klog.Errorf("no result found")
		os.Exit(1)
	}

	var versionString string
	if err := version.Scan(&versionString); err != nil {
		klog.Errorf("cannot use string as return variable: %s\n")
		os.Exit(1)
	}

	klog.Infof("database version string: %s", versionString)
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

	db, err := sql.Open("postgres", conStr)
	if err != nil {
		klog.Errorf("cannot connect to db: %s", err)
		os.Exit(1)
	}
	klog.Infof("connect to database")

	dbVersion(db)

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

	authHelper := auth.NewAuthHelper(db, "contract_create")

	go authHelper.CleanUp()

	authHandler, err := auth.NewOidcAuth(conf.UserMgmt.UserMgmt, conf.UserMgmt.UserRealm, conf.UserMgmt.BasePath, pas.UserMgmt.ClientSecret, pas.UserMgmt.ClientId, conf.UserMgmt.ServerAddress,  authHelper)
	if err != nil {
		klog.Errorf("cannot create new oidc handler: %s", err)
		os.Exit(1)
	}

	contractMachineDataHandler := machineData.NewPsqlContract(db)

	klog.Infof("define endpoints")
	machineHandler := machineData.NewMachineDataEndpoint(sendChan, authHelper, contractMachineDataHandler)

	analysisHandler := analysisModel.NewAnalysisHandler(db)
	analysisResultListHandler := analysisModel.NewResultList(db)
	analysisLogic := analysis.NewAnalyseLogic(analysisResultListHandler, analysisHandler)
	analysisEndpoint := analysis.NewAnalysisEndpoint(analysisLogic, authHelper)

	contractHandleWorker := contractModel.NewContractHandler(db, "cloud")
	contractResultList := contractModel.NewResultList(db)
	contractLogic := contract.NewContractLogic(contractResultList, contractHandleWorker, "cloud")
	contractHandler := contract.NewContractEndpoint(contractLogic, authHelper)

	http.Handle("/auth", authHandler)
	http.Handle("/auth/", authHandler)
	http.Handle("/health", new(health.Health))
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/machine-data", machineHandler)
	http.Handle("/ready", new(ready.Ready))
	http.Handle("/analysis/", analysisEndpoint)
	http.Handle("/contract/", contractHandler)

	//http.Handle("/analyses/", analysesResult)
	//http.Handle("/model/", model)

	klog.Infof("start webserver")
	listen := fmt.Sprintf("%s:%d", conf.Webserver.Address, conf.Webserver.Port)
	_ = http.ListenAndServe(listen, nil)
}
