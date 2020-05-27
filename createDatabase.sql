CREATE TABLE IF NOT EXISTS token (token uuid PRIMARY KEY, name TEXT);
CREATE TABLE IF NOT EXISTS contract (id TEXT PRIMARY KEY, delete bool DEFAULT false);
CREATE TABLE IF NOT EXISTS model (id BIGSERIAL PRIMARY KEY, tag TEXT, url TEXT);
CREATE TABLE IF NOT EXISTS model_cloud(model BIGINT REFERENCES model, contract TEXT REFERENCES contract);
CREATE TABLE IF NOT EXISTS model_edge(model BIGINT REFERENCES model, contract TEXT REFERENCES contract);
CREATE TABLE IF NOT EXISTS machine(id TEXT PRIMARY KEY);
CREATE TABLE IF NOT EXISTS sensor(id BIGSERIAL PRIMARY KEY, transmitted_id TEXT);
CREATE TABLE IF NOT EXISTS machine_contract(machine TEXT REFERENCES machine(id), contract TEXT REFERENCES contract);
CREATE TABLE IF NOT EXISTS machine_sensor(id BIGSERIAL PRIMARY KEY, machine TEXT REFERENCES  machine(id), sensor BIGINT REFERENCES sensor);
CREATE TABLE IF NOT EXISTS sensor_model(sensor BIGINT REFERENCES machine_sensor, model BIGINT REFERENCES model);
CREATE TABLE IF NOT EXISTS analyse_result(ID BIGSERIAL PRIMARY KEY, machine TEXT REFERENCES machine, sensor TEXT, time TIMESTAMP, result JSON, contract TEXT REFERENCES contract(id));
