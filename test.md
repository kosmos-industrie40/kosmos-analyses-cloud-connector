# Test

In this file, you can find a description, how to test different endpoints.

## Content

1. [Auth](#auth)
1. [Contracts](#contracts)
1. [Upload Sensor Data](#upload-sensor-data)
1. [Analyses](#analyse-results)
1. [Metrics](#metrics)
1. [Model](#model)

## Auth
In this chapter a simple test case against the auth endpoint can be found. In three steps we will try to logged in, test if we are already logged in and log out.
We are using `curl` to send the requests.

### Log in
```bash
curl -i -X POST localhost:8080/auth --data '{"user":"test", "password":"abc"}'
```

### Test Authenticated
```bash
curl -i --header 'token:(RETURN VALUE FROM LOG IN REQUEST)' localhost:8080/auth
```

### Logout
```bash
curl -i -X DELETE --header 'token:(RETURN VALUE FROM LOG IN REQUEST)' localhost:8080/auth
```

## Contracts

Before the following queries could be used, you have to log in to the api.
You can use the Auth/Login query from the previous chapter.

### Upload Contract
```
curl -X POST --header 'token:<insert auth token here>' -i localhost:8080/contract/ --data @testContract.json
```
where `insert auth token here` is the authentication token from the login.

### List of all Contracts
```
curl --header 'token:<insert auth token here>' -i localhost:8080/contract/
```
where `insert auth token here` is the authentication token from the login.

### List specific Contract
```
curl --header 'token:<insert auth token here>' -i localhost:8080/contract/<contractId>
```
where `insert auth token here` is the authentication token from the login.
where `contractId` is the specific contract.

### Delete Contract
```
curl -X DELETE --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' -i localhost:8080/contract/<contractId>
```

where `insert auth token here` is the authentication token from the login.
where `contractId` is the specific contract.

## Upload Sensor Data
You have to log in before you can upload your data. To do this, you can use the described way in [Auth/Log in](#auth_log_in).

You can find an example of sensor data, which can be uploaded in the `exampleData.json` file.
To view the data output you have to start a mqtt subscriber. You can use `mosquitto_sub` from
from the mosquitto-clients package. To start this, you can use this command `mosquitto_sub -t 'kosmos/machine-data/#'`

To upload data you can use the following exmple request.
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
In the last section we want to set a stae to the specific contract. Before you can execute the following command, you should
inserted the data into the database.

```bash
curl -X PUT --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/model/contract-test33 --data '{"state":"test", "models":[{"tag":"tag", "url":"url"}]}'
```
