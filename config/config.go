package config

import (
	"io"
	"os"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/klog"
)

//TODO https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/-/issues/3
func ParseConfiguration(path string, conf interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			klog.Errorf("could not close file; err: %v", err)
		}
	}()

	return handleConfiguration(file, conf)
}

func handleConfiguration(handle io.Reader, conf interface{}) error {
	decoder := yaml.NewDecoder(handle)
	decoder.SetStrict(true)

	if err := decoder.Decode(conf); err != nil {
		return err
	}

	return nil
}
