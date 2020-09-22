package models

type Contract struct {
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
		Name            string `json:"name"`
		StorageDuration []struct {
			SystemName string `json:"systemName"`
			Duration   string `json:"duration"`
		} `json:"storageDuration"`
		Meta interface{} `json:"meta,omitempty"`
	} `json:"sensors"`
	CheckSignature bool `json:"checkSignatures"`
	Analysis       struct {
		Enable  bool `json:"enable"`
		Systems []struct {
			Enable    bool   `json:"enable"`
			Name      string `json:"system"`
			Pipelines []struct {
				Trigger struct {
					Type       string             `json:"type"`
					Definition *TriggerDefinition `json:"definition"`
				} `json:"ml-trigger"`
				Pipeline []struct {
					Container Container `json:"container"`
					Persist   bool      `json:"persistOutput"`
					From      *Model    `json:"from"`
					To        *Model    `json:"to"`
				} `json:"pipeline"`
				Sensors []string `json:"sensors"`
			} `json:"pipelines"`
		} `json:"systems"`
	} `json:"analysis"`
	Metadata interface{} `json:"metadata"`
	MachineConnection interface{} `json:"machineConnection"`
	Blockchain interface{} `json:"blockchain"`
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
