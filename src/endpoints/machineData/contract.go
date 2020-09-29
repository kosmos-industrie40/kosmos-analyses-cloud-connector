package machineData

import (
	"database/sql"

	"k8s.io/klog"
)

type Contract interface {
	GetContracts(machine, sensor string) ([]string, error)
}

func NewPsqlContract(db *sql.DB) Contract {
	return psqlContract{db: db}
}

type psqlContract struct {
	db *sql.DB
}

func (p psqlContract) GetContracts(machine, sensor string) ([]string, error) {
	query, err := p.db.Query("SELECT contract FROM contract_machine_sensors AS cms JOIN machine_sensors ms on cms.machine_sensor = ms.id JOIN sensors s on ms.sensor = s.id WHERE machine = $1 AND sensor = $2", machine, sensor)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	var contracts []string
	for query.Next() {
		var contract string
		if err := query.Scan(&contract); err != nil {
			return nil, err
		}

		contracts = append(contracts, contract)
	}
	
	return contracts, nil
}

