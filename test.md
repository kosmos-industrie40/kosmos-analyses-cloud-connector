# Test

This file contains a description how the single endpoints could be tested.

## Content
1. [Dependencies](#dependencies)
1. [Requirements](#requirements)
1. [Authentication](#authentication)
1. [Contracts](#contracts)
1. [Upload Sensor Data](#upload-sensor-data)
1. [Analyses](#analyse-results)
1. [Metrics](#metrics)
1. [Model](#model)

## Dependencies
To execute all of theses test we are using the command line tools:
- curl
- mosquitto\_sub
- jq
- psql

On a Linux system you can install them via your package manger.
The following example using a Debian based Linux distribution.
```bash
apt update
apt upgrade -y
apt install -y curl mosquitto-clients jq postgresql-client
```

## Requirements
In the most endpoints a header file with a token has to be added. To simplify the commands we can use a `test-user`
with a predefined token. This will be inserted to the database and you can use them in the test cases.

To insert these data point you can use this command:
```bash
psql -h <host> -d <database> -U <database user> -c \
"INSERT INTO token (name, token) VALUES ('test-user', 'ca397616-e351-47c3-ae7b-0785e6278357');"
```

In this test we are assuming that the program is running on the `localhost` and using the port `8080`.

The test data files are located in the `test` directory. In this test file we are assuming, that the working directory
is in the test directory.

## Authentication
This section contains the test case to test against the authentication endpoint. 

### Log in
```bash
curl -i -X POST localhost:8080/auth --data '{"user":"test", "password":"abc"}'
```

### Test Authenticated
```bash
curl -i --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/auth
```

### Logout
```bash
curl -i -X DELETE --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/auth
```
(In this case you are delete the token and if you want to use the example token in the next steps you have
to reinsert the user-token combination)

## Contracts
In this section the contract endpoint will be tested

### Upload Contract
```
curl -X POST --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' -i localhost:8080/contract/ --data @testContract.json
```

### List of all Contracts
```
curl --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' -i localhost:8080/contract/
```

### List specific Contract
```
curl --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' -i localhost:8080/contract/<contractId>
```
Where `contractId` is the specific contract.

### Delete Contract
```
curl -X DELETE --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' -i localhost:8080/contract/<contractId>
```

Where `contractId` is the specific contract.

## Upload Sensor Data
You can find an example of sensor data, which can be uploaded in the `exampleData.json` file.
To view the data output you have to start a mqtt subscriber. The following command can be used to
view messages on a mqtt topic.
```bash
mosquitto_sub -t 'kosmos/machine-data/#'
```

To upload data you can use the following example request.
```bash
 curl -i -X POST --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/machine-data/ --data @exampleData.json
```
## Analyse Results
in this chapter is the description how to test the analyses result endpoint. The 
endpoint is divided into three parts.

### Get Specific Result
To receive a specific analyse result you can use the following curl command. 
```bash
curl -i --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/analyses/77/8
```

### Get Result Set
To receive all results from a specific contract:
```bash
curl -i --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/analyses/77
```


### Upload Result
Before you can upload the data you have the contract and the machine has to be created.
```bash
curl -i -X POST --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/analyses/77/mach1/sens1 --data @exampleAnalyseResult.json
```

## Metrics
The metric endpoint provides prometheus metrics which are created from the promehteus golang client. So you can query those with the following
command:
```bash
curl -i localhost:8080/metrics
```

## Model
The model endpoint provides the functionality to the model endpoint. On this you can query for model updates and update
the state of the model.

### Get Model
This endpoint uses a get request to query the required models.
```bash
curl -i --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/model/77
```
The last request should be empty. Before we can query data; we have to insert those data into the database. With the next
code block contains the data which should be inserted.

```bash
psql -d <database> -c "INSERT INTO machine VALUES ('contract')"
psql -d <database> -c "INSERT INTO model (id, tag, url) VALUES (0, 'tag', 'url')"
psql -d <database> -c "INSERT INTO model_update (model, contract) VALUES (0, 'contract-test33')"
```

```bash
curl --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/model/contract-test33
```

### Update model
In the last section we want to set a state to the specific contract. Before you can execute the following command, you should
inserted the data into the database. The command to insert those can be found in the previous section.

```bash
curl -X PUT --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/model/contract-test33 --data '{"state":"test", "models":[{"tag":"tag", "url":"url"}]}'
```
