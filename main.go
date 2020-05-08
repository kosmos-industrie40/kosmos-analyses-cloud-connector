package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog"

	"flag"
	"fmt"
	"net/http"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/config"
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

	config.ParseConfiguration(cli.configuration, &conf)
	config.ParseConfiguration(cli.password, &pas)

	klog.Infof("configuration is parsed")

	// paths
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/analyses/", new(endpoints.Analyses))
	http.Handle("/machine-data/", new(endpoints.MachineData))
	http.Handle("/auth", new(endpoints.Auth))
	http.Handle("/health", new(endpoints.Health))
	http.Handle("/model/", new(endpoints.Model))
	http.Handle("/contract/", new(endpoints.Ready))

	klog.Infof("start webserver")
	listen := fmt.Sprintf("%s:%d", conf.Webserver.Address, conf.Webserver.Port)
	http.ListenAndServe(listen, nil)
}
