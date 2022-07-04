package models

import (
	"time"

	"k8s.io/klog"
)

type StorageDuration struct {
	SystemName string `json:"systemName"`
	Duration   string `json:"duration"`
}

type Contract struct {
	Body struct {
		Contract struct {
			Valid struct {
				Start string `json:"start"`
				End   string `json:"end"`
			} `json:"valid"`
			CreationTime string   `json:"creationTime"`
			Partners     []string `json:"partners"`
			Permissions  struct {
				Read  []string `json:"read"`
				Write []string `json:"write"`
			} `json:"Permissions"`
			ID      string `json:"id"`
			Version string `json:"version"`
		} `json:"contract"`
		TechnicalContainers []struct {
			System     string      `json:"system"`
			Containers []Container `json:"Containers"`
		} `json:"requiredTechnicalContainers"`
		Machine            string   `json:"machine"`
		KosmosLocalSystems []string `json:"kosmosLocalSystems"`
		Sensors            []struct {
			Name            string          `json:"name"`
			StorageDuration []StorageDuration `json:"storageDuration"`
			Meta            interface{}     `json:"meta,omitempty"`
		} `json:"sensors"`
		CheckSignature    bool        `json:"checkSignatures"`
		Analysis          Analysis    `json:"analysis"`
		Metadata          interface{} `json:"metadata"`
		MachineConnection interface{} `json:"machineConnection"`
		Blockchain        interface{} `json:"blockchain"`
	} `json:"body"`
	// TODO update signature to new object
	Signature interface{} `json:"signature"`
}

type Analysis struct {
	Enable  bool `json:"enable"`
	Systems []struct {
		Enable    bool   `json:"enable"`
		Name      string `json:"system"`
		Pipelines []struct {
			Trigger  Trigger    `json:"ml-trigger"`
			Pipeline []Pipeline `json:"pipeline"`
			Sensors  []string   `json:"sensors"`
		} `json:"pipelines"`
	} `json:"systems"`
}

type Trigger struct {
	Type       string             `json:"type"`
	Definition *TriggerDefinition `json:"definition"`
}

type Pipeline struct {
	Container Container `json:"container"`
	Persist   bool      `json:"persistOutput"`
	From      *Model    `json:"from"`
	To        *Model    `json:"to"`
}

// TODO validate the message against the signature
// TODO ignore other remote systems
func (c Contract) Valid(localSystem string) bool {
	body := c.Body
	systems := make(map[string]bool)

	for _, system := range body.KosmosLocalSystems {
		systems[system] = true
	}

	systems[localSystem] = true

	for _, analyse := range body.Analysis.Systems {
		if _, ok := systems[analyse.Name]; !ok {
			klog.Infof("analyse.Name %s doesn't exists", analyse.Name)
			return false
		}
	}

	for _, container := range body.TechnicalContainers {
		if _, ok := systems[container.System]; !ok {
			return false
		}
	}

	start, err := time.Parse(time.RFC3339, body.Contract.Valid.Start)
	if err != nil {
		klog.Errorf("cannot parse start validation: %s", err)
		return false
	}

	end, err := time.Parse(time.RFC3339, body.Contract.Valid.End)
	if err != nil {
		klog.Errorf("cannot parse end validation: %s", err)
		return false
	}

	if start.After(end) {
		return false
	}

	_, err = time.Parse(time.RFC3339, body.Contract.CreationTime)
	if err != nil {
		klog.Errorf("cannot parse contract creation time: %s", err)
		return false
	}

	return true

}

type Model struct {
	Tag string `json:"tag"`
	Url string `json:"url"`
}

type Container struct {
	Url         string   `json:"url"`
	Tag         string   `json:"tag"`
	Arguments   []string `json:"arguments"`
	Environment []string `json:"environment"`
}

type TriggerDefinition struct {
	After string `json:"after"`
}
