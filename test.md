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

On a Linux system you can install them via your package manager.
The following example shows how to install them using a Debian based Linux distribution.
```bash
apt update
apt upgrade -y
apt install -y curl mosquitto-clients jq postgresql-client
```

## Requirements
For most endpoints a header file with a token has to be added. To simplify the commands we can use a `test-user`
with a predefined token. This will be inserted to the database and you can use them in the test cases.

To insert these data point you can use following commands:
```bash
psql -h <host> -d <database> -U <database user> -c \
"INSERT INTO token(token, valid, write_contract) VALUES ('ca397616-e351-47c3-ae7b-0785e6278357', NOW() + '5h', 't');"
psql -h <host> -d <database> -U <database user> -c \
"INSERT INTO organisations(id, name) VALUES (0, 'test');"
psql -h <host> -d <database> -U <database user> -c \
"INSERT INTO token_permission(token, organisation) VALUES ('ca397616-e351-47c3-ae7b-0785e6278357', 0);"
```

The following commands assume that the mediator is running on `localhost` and uses the port `8080`. 

The test data files are located in the `test` directory. Therefore we are assuming, that your working directory is the test directory.

Switch to the test directory

```bash
cd test/
```

## Authentication
This section specifies how to test against the authentification endpoint.

### Log in
To test this endpoint, you have to set up or use a oidc auth server. This can be a self hosted instance of keycloak, google auth or github. The best way to test this is by using a browser and
call the endpoint with the 'auth' path.

### Test Authenticated
```bash
curl -i --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/auth
```

### Logout
```bash
curl -i -X DELETE --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/auth
```
**Warning**: This deletes the token from the database. For the following further steps it is necessary to have a valid token! Therefore you have to reinsert the user-token combination again.


## Contracts
In this section the contract endpoint will be tested

Contracts are a central component of communication in KOSMoS. All queries, calculations and actions depend to a large extent on whether a valid contract exists and what is mapped with this contract. Thus, most components also depend on the specification of the contracts. 

In the further course we will use the ``valid_example.json`` contract. Key Information of this contract:

- ```id```: 53
- ```read/write permission``` for organisation test
- ```machine```: 84bab968-e6b7-11ea-b10c-54e1ad207114
- ```sensors``` : temperature , alarms , crash


###  <a name="upload_contract"></a> Upload Contract


```bash
curl -X POST --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' -i localhost:8080/contract/ --data valid_example.json
```

Upper curl request creates a contract with ID 53 that can be read/written by the ```test``` organization. 

### List of all Contracts
```bash
curl --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' -i localhost:8080/contract/
```

### List specific Contract
```bash
curl --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' -i localhost:8080/contract/<contractId>
```
Where `contractId` is the specific contract.

### Delete Contract
```bash
curl -X DELETE --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' -i localhost:8080/contract/<contractId>
```

Where `contractId` is the specific contract.

**Important Note**: Calling the delete request will only cause the ```active``` attritbute to be set to ```false```. The contract is still in the database and is still displayed in the list of all contracts. 

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

After running above command you should see following message on the MQTT subscriber:


```json
{
   "body":{
      "machineID":"",
      "timestamp":"2020-08-15T15:33:44.897Z",
      "columns":[
         {
            "name":"value",
            "type":"number",
            "meta":{
               "unit":"",
               "description":"metric"
            }
         },
         {
            "name":"quality",
            "type":"number",
            "meta":{
               "unit":"",
               "description":"quality of the metric"
            }
         }
      ],
      "data":[
         [
            "15"
         ],
         [
            "3"
         ]
      ],
      "meta":null
   },
   "signature":""
}
```

## Analysis Results
in this chapter is the description how to test the analyses result endpoint. The 
endpoint is divided into three parts.

### Upload Result
Before you can upload the data, the contract and the machine has to be created. The best way to do this, is to upload a contract first (See [Upload Contract](#a-name"uploadcontract"a-upload-contract)). After a contract is uploaded you can upload a result with this command
```bash
curl -i -X POST --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/analysis/53/84bab968-e6b7-11ea-b10c-54e1ad207114/temperature --data @exampleAnalyseResult.json
```

The endpoint general structure is ```localhost:8080/analysis/<contractID>/<machineID>/<sensorID>```

### Get Result Set
To receive all results from a specific contract:
```bash
curl -i --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/analysis/53
```

The endpoint general structure is ```localhost:8080/analysis/<contractID>```

If you uploaded a contract you should get a list like below:

```bash
[{"resultID":1,"machine":"84bab968-e6b7-11ea-b10c-54e1ad207114","date":"2020-08-12T17:46:10.821+02:00"}]
```


### Get Specific Result
To receive a specific analyse result you can use the following curl command. 
```bash
curl -i --header 'token:ca397616-e351-47c3-ae7b-0785e6278357' localhost:8080/analysis/53/1
```


The endpoint general structure is ```localhost:8080/analysis/<contractID>/<resultID>/```

For our example you get a json like following:

```json
{
   "body":{
      "from":"creator of this message",
      "timestamp":"2020-08-12T15:46:10.821Z",
      "model":{
         "url":"abc",
         "tag":"ab"
      },
      "type":"text",
      "calculated":{
         "message":{
            "machine":"abc",
            "sensor":"134wdsf"
         },
         "received":"2020-08-12T15:47:10.821Z"
      },
      "results":{
         "parts":[
            {
               "machine":"machine1",
               "predict":90,
               "result":"stop",
               "sensors":[
                  {
                     "predict":100,
                     "result":"stop",
                     "sensor":"sensor1"
                  }
               ]
            }
         ],
         "predict":80,
         "total":"stop"
      }
   },
   "signature":""
}
```



## Metrics
The metric endpoint provides prometheus metrics which are created from the promehteus golang client. So you can query those with the following
command:
```bash
curl -i localhost:8080/metrics
```
