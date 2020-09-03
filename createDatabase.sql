BEGIN;
CREATE TABLE IF NOT EXISTS systems (
	id bigserial PRIMARY KEY,
	name text UNIQUE
);

CREATE TABLE IF NOT EXISTS contracts (
	id text PRIMARY KEY,
	start_time timestamptz,
	end_time timestamptz,
	creation timestamptz,
	validate_signature boolean,
	metadata json,
	machineConnection json,
	blockchain json
);

CREATE TABLE IF NOT EXISTS organisations (
	id bigserial PRIMARY KEY,
	name text UNIQUE
);

CREATE TABLE IF NOT EXISTS kosmos_local (
	contract text REFERENCES contracts,
	system bigint REFERENCES systems
);

CREATE TABLE IF NOT EXISTS containers (
	id bigserial PRIMARY KEY,
	url text,
	tag text,
	arguments text[],
	environment text[]
);

CREATE TABLE IF NOT EXISTS connection (
	system bigint REFERENCES systems,
	interval interval,
	url text,
	user_mgmt text,
	container bigint REFERENCES containers
);

CREATE TABLE IF NOT EXISTS partners (
	contract text REFERENCES contracts,
	organisation bigint REFERENCES organisations
);


CREATE TABLE IF NOT EXISTS sensors (
	id bigint PRIMARY KEY,
	transmitted_id text,
	meta json
);

CREATE TABLE IF NOT EXISTS machines (
	id text PRIMARY KEY,
	meta json
);

CREATE TABLE IF NOT EXISTS machine_sensors (
	id bigserial PRIMARY KEY,
	machine text REFERENCES machines,
	sensor bigint REFERENCES sensors
);

CREATE TABLE IF NOT EXISTS contract_machine_sensors (
	id bigserial PRIMARY KEY,
	contract text REFERENCES contracts,
	machine_sensor bigint REFERENCES machine_sensors
);

CREATE TABLE IF NOT EXISTS storage_duration (
	system bigint REFERENCES systems,
	contract_machine_sensor bigint REFERENCES contract_machine_sensors,
	duration interval
);


CREATE TABLE IF NOT EXISTS analyse_result (
	id bigint,
	contract_machine_sensor bigint REFERENCES contract_machine_sensors,
	time timestamptz,
	result json,
	status text
);

CREATE TABLE IF NOT EXISTS models (
	id bigint PRIMARY KEY,
	container bigint REFERENCES containers
);

CREATE TABLE IF NOT EXISTS analyses (
	id bigserial PRIMARY KEY,
	contract_machine_sensor bigint REFERENCES contract_machine_sensors,
	prev_model bigint REFERENCES models,
	next_model bigint REFERENCES models
);

CREATE TABLE IF NOT EXISTS pipelines (
	id bigserial PRIMARY KEY,
	contract text REFERENCES contracts,
	ana bigint REFERENCES analyses,
	system bigint REFERENCES systems,
	event_tigger boolean,
	time_trigger interval
);

CREATE TABLE IF NOT EXISTS update_message (
	machine_sensor integer REFERENCES machine_sensors,
	time timestamp,
	meta json,
	columns json,
	data json
);

CREATE TABLE IF NOT EXISTS technical_containers (
	contract text REFERENCES contracts,
	container bigint REFERENCES containers,
	system bigint REFERENCES systems
);

CREATE TABLE IF NOT EXISTS write_permissions {
	contract TEXT REFERENCES contracts,
	organisation BIGINT REFERENCES organistsions
}

CREATE TABLE IF NOT EXISTS read_permissions {
	contract TEXT REFERENCES contracts,
	organisation BIGINT REFERENCES organistsions
}

COMMIT;
