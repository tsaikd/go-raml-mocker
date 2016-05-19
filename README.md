# go-raml-mocker
RAML (1.0) web mock server implementation in golang

[![Build Status](https://travis-ci.org/tsaikd/go-raml-mocker.svg?branch=master)](https://travis-ci.org/tsaikd/go-raml-mocker)

## Features

* Live reload web mock server routes from RAML file

## Install

```
go get -u -v "github.com/tsaikd/go-raml-mocker"
```

## Usage

* start mock web server

```
go-raml-mocker -f example/organisation-api.raml --port 4000
```

* try to get data from mock web server

```
curl http://localhost:4000/organisation
```
