# configurator

A Golang utility for setting up application configuration

---

[![Build Status](https://travis-ci.org/marksost/configurator.svg?branch=master)](https://travis-ci.org/marksost/configurator) [![Current Release](https://img.shields.io/badge/release-0.1.1-1eb0fc.svg)](https://github.com/marksost/configurator/releases/tag/0.1.1) [![GoDoc](https://godoc.org/github.com/marksost/configurator?status.svg)](https://godoc.org/github.com/marksost/configurator)

## Installation and Usage

You can install this package locally with:

```go
go get github.com/marksost/configurator
```

and then import it into your project with:

```go
import "github.com/marksost/configurator"
```

and then use it by passing in your configuration struct like:

```go
// Define your config
type Config struct { /* ...your config properties go here... */ }

// Create a new config instance
c := &Config{}

// Initialize it!
configurator.InitializeConfig(c)
```
