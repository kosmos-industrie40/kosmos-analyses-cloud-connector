# Connector Analyses Cloud - Edge
This repository contains the endpoint definition and the implementation of the
connector between edge and analyses platform. This program provides different
endpoints, where you can execute actions on the analyse result, kosmos contracts,
sensor data, machine learning models.

To download the full project please use `git clone --recursive`. To execute all test cases, we are using a git submodule. To update the submodule you can run `git submodule update --init --recursive`

## Content

- [Connector Analyses Cloud - Edge](#connector-analyses-cloud---edge)
	- [Content](#content)
	- [Endpoint Definition](#endpoint-definition)
	- [Dependencies](#dependencies)
	- [Build](#build)
	- [Test](#test)
	- [Configuration](#configuration)
		- [CLI-Flags](#cli-flags)
		- [Password](#password)
		- [Configuration](#configuration-1)

## Endpoint Definition

The endpoint definition is created as openapi 3.0.1 document.
General informations about openapi can be found on [this swagger page.](https://swagger.io/docs/specification/about/)
To view in the api definition please open the [ConnectorEdgeCloud.yaml file.](./ConnectorEdgeCloud.yaml)

## Dependencies
Golang 1.15 is used to write this endpoint. So golang is 
one of the requirements. We are using go modules to organize the sufficient dependencies. Those
are organized in the `go.mod` file.

There are a few extra infrastructure dependencies. You need to set up a PostgreSQL database server 
and a MQTT-Server. To install PostgreSQL check out [Download PostgresSQL page](https://www.postgresql.org/download/). 
As MQTT-Broker you can use Mosquitto from the eclipse foundation. To deploy
or install Mosquitto check out [Download Mosquitto page](https://mosquitto.org/download/)

## Build
You can build this program by executing `make` or `go build ./...`. 

Before you can execute this program you should create the database layout.
You can use the file `createDatebase.sql` to create the required Tables.
The following command gives an example to create the database tables.
```bash
psql -h <host> -d <database> -U <database user>  < createDatebase.sql
```
The following variables has to be set to you specific deployment:
- host
- database
- database

The resulting database schema looks as follows:
![KOSMoS Analyses DB Schema](kosmos-analyse-db.jpeg)

To start the programm run

```
./connector
```

## Test
We have created an extra file, on which all the endpoints are checked by using extra commands. Please checkout
the [test file](test.md).

## Configuration
The configuration of the application will be made through two configuration files and command line flags. 
The configuration parameters will be explained in the next three sections.

### CLI-Flags
In this section the command line parameters will be displayed. Flags which are created by the logging tool `klog` will not be
acknowledge in this chapter.

| flag | default value | description |
|------|---------------|-------------|
| pass | examplePassword.yaml | is the path to the password configuration file |
| config | exampleConfig.yaml | is the path to the configuration file |

### Password
The password configuration contains passwords for the database connection and the mqtt connection. An example can be
found in the `examplePassword.yaml` file.

|parameter|description|
| ------- | --------- |
| mqtt.user | is the user name of the mqtt user which is used for the mqtt connection |
| mqtt.password | is the password which is used by the mqtt.user for the mqtt connection |
| database.user | is the user for the postgresql database connection |
| database.password | is the password for the postgresql database connection |
| userMgmt.clientID | is the client id, which is used by the user management |
| userMgmt.clientSecret | is the client secret, which is used by the user management |

### Configuration
The configuration file will be used to configure the system without including credentials. An example configuration
can be found in the `exampleConfiguration.yaml` file.

| parameter | description |
| --------- | ----------- |
| webserver.address | is the IP address on which this application will be open the web server|
| webserver.port | is the port this application used for the web server |
| database.address | is the IP address (or URL), where the PostgreSQL server could be found |
| database.port | is the port of the PostgreSQL server |
| database.database | is the name of the PostgreSQL database |
| mqtt.address | is the IP address (or URL) of the mqtt broker |
| mqtt.port | is the port of the mqtt broker|
| userMgmt.userMgmt | is the address to the user managment system (on keycloak inclusive realm) |
| userMgmt.serverAddress | is the local server address |
