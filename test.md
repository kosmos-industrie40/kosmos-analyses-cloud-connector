# Test

In this file, you can find a description, how to test different endpoints.

## Content

1. [Auth](#auth)
1. [Contracts](#contract)
1. [Upload Sensor Data](#upload_sensor_data)

## Auth
In this chapter a simple test case against the auth endpoint can be found. In three steps we will try to logged in, test if we are already logged in and log out.
We are using `curl` to send the requests.

### Log in
```bash
curl -i -X POST localhost:8080 --data '{"user":"test", "password":"abc"}'
```

### Test Authenticated
```bash
curl -i --header 'token:(RETURN VALUE FROM LOG IN REQUEST)' localhost:8080 
```

### Logout
```bash
curl -i -X DELETE --header 'token:(RETURN VALUE FROM LOG IN REQUEST)' localhost:8080 
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
