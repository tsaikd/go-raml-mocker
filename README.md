# go-raml-mocker [![Build Status](https://travis-ci.org/tsaikd/go-raml-mocker.svg?branch=master)](https://travis-ci.org/tsaikd/go-raml-mocker) [![Report card](https://goreportcard.com/badge/github.com/tsaikd/go-raml-mocker)](https://goreportcard.com/report/github.com/tsaikd/go-raml-mocker)

RAML (1.0) web mock server implementation in golang

## Features

* Live reload web mock server routes from RAML file

## Use pre-build binary from docker hub

* start mock web server

```
docker run \
	-p 4000:4000 \
	-v "${PWD}/example/organisation-api.raml:/raml/organisation-api.raml" \
	tsaikd/go-raml-mocker:1.0 \
	go-raml-mocker -f /raml/organisation-api.raml --port 4000
```

* try to get data from mock web server

```
curl http://localhost:4000/organisation
```

## Use golang binary from github source code

### Install

```
go get -u -v "github.com/tsaikd/go-raml-mocker"
```

### Usage

* start mock web server


```
go-raml-mocker -f example/organisation-api.raml --port 4000
```

* try to get data from mock web server

```
curl http://localhost:4000/organisation
```
