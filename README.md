# Connector Analyses Cloud - Edge

This repository contains the api definition, and the implemented endpoint of
the analyses cloud connector. This program can be used to push data into the 
analyse cloud, query results of the analyses cloud and ask for model updates.

## Content

- [Api Definition](#api-definition)
- [Dependencies](#dependencies)
- [Build](#build)
- [Configuration](#configuration)

## Api Definition

The api definition is created as openapi 3.0.1 document.
General informations about openapi can be found on [this swagger page.](https://swagger.io/docs/specification/about/)
To view in the api definition please open the [ConnectorEdgeCloud.yaml file.](./ConnectorEdgeCloud.yaml)

## Dependencies
We are using golang in the version 1.14. to start this programm. So golang is 
one of the requirements. We are using go modules to organize the suffizient dependencies. Those
are organized in the file `go.mod`.

There are a few extra infrasturcter dependencies. You need to set up a PostgreSQL database server 
and a MQTT-Server. To install PostgreSQL check out [Download PostgresSQL page](https://www.postgresql.org/download/). As MQTT-Broker you can use Mosquitto from the eclipse foundation. To deploy
or install Mosquitto check out [Download Mosquitto page](https://mosquitto.org/download/)

## Build
You can build this programm by executing `make` or `go build ./...`. To test the functionality 
check out the [test.md file](test.md).

## Configuration
To configure the web application there are two importend cli-flags and two configuration files.
We are using two files to sepearte the user/password from the other configuration.

### CLI
| flag | default value | description |
|------|---------------|-------------|
| pass | examplePassword.yaml | is the path to the password configuration file |
| config | exampleConfig.yaml | is the path to the configuration file |

### Password
TODO

### Configuration
TODO
