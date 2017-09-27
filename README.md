# configurator

A Go utility for setting up application configuration

---

[![Build Status](https://travis-ci.org/marksost/configurator.svg?branch=master)](https://travis-ci.org/marksost/configurator) [![Current Release](https://img.shields.io/badge/release-0.2.0-1eb0fc.svg)](https://github.com/marksost/configurator/releases/tag/0.2.0) [![GoDoc](https://godoc.org/github.com/marksost/configurator?status.svg)](https://godoc.org/github.com/marksost/configurator)

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

## Explanation of functionality

Configurator is designed to take in an interface value - a configuration struct - which is used for configuration for a Go application. This stuct can contain any properties it wants, but each should have a set of tags that tell configurator how to set it's value. Currently, it supports the following types for properties: `string`, `int`, `boolean`, `struct`. NOTE: by supporting `struct`s, configurator will recursively set properties of that struct, which makes setting up logical configuration groups within the global object easy. Any type not listed will be ignored during the below operations (but adding support for types is welcomed via PR!).

Each property of the configuration struct should have three tags: `default`, `json`, and `env`. In order, these tags provide a default value for the property (NOTE: this allows you to set a different "zero-value", so to speak, for a property), a field label that a JSON configuration file could be unmarshaled into, and finally an environment variable name that, if present, will be used as the property's value.

An example configuration struct you might pass into configurator looks like:

```go
type (
  Example struct {
    Property1 string `default:"default-value" json:"property-one" env:"PROPERTY_ONE"`
    Property2 int `default:"1234" json:"property-two" env:"PROP_TWO"`
    Property3 bool `default:"true" json:"property-three" env:"DIFFERENT_ENV_VARIABLE_NAME"`
    Property4 StructPropertyExample `json:"property-four"` // NOTE: Provided for JSON unmarshaling when possible
  }
  StructPropertyExample struct {
    SubProp1 string `default:"foo" json:"sub-prop-1" env:"SUB_PROP_ONE"`
  }
)
```

When this struct is passed into `configurator.InitializeConfig`, a number of operations are applied to it. First, configurator loops through each feild and attempts to set the property's value to that which is in the `default` tag. For example, after this pass, the above configuration struct would look like:

```go
Example{
  Property1: "default-value",
  Property2: 1234,
  Property3: true,
  Property4: StructPropertyExample{
    SubProp1: "foo",
  },
}
```

After default values are set, configurator will attempt to load a valid JSON configuration file. The full path (including file name) should be set in your environment under a variable name that matches `configurator.ConfigLocation`. This variable can be changed by importing applications, and should be set before calling `InitializeConfig`.

If the environment variable doesn't exist, the file it points to doesn't exist (or can't be read), or the file contains invalid JSON, the workflow moves on to the next step. Otherwise, the contents of the file are unmarshaled into the configuration struct, setting any properties that have matching JSON field labels. For example, if the following JSON configuration file was found:

```json
{
  "property-one": "json-value",
  "property-three": false,
  "property-four": {
    "sub-prop-1": "bar"
  }
}
```

after this pass, the above configuration struct would look like:

```go
Example{
  Property1: "json-value",
  Property2: 1234,
  Property3: false,
  Property4: StructPropertyExample{
    SubProp1: "bar",
  },
}
```

NOTE: Configuration files like this are not normally used in non-development environments, but are a convenient way to develop locally without having to set a ton of environment variables.

After the JSON configuration file is applied, configurator will attempt to read in environment variables that match each configuration struct property's `env` tag (see below for an explanation on environment variable name formation). For example, if the following environment variables were set, and the value of `configurator.EnvPrefix` was `TEST_`:

```bash
export TEST_PROP_TWO=2345
export TEST_SUB_PROP_ONE="baz"
```

after this pass, the above configuration struct would look like:

```go
Example{
  Property1: "json-value",
  Property2: 2345,
  Property3: false,
  Property4: StructPropertyExample{
    SubProp1: "baz",
  },
}
```

Finally, for each `env` tag found and formatted, a command-line flag will be added to the `flags` package (see below for an explanation on flag name formation). After all environment variables are processed, configurator will call `flags.Parse()`, which allows for one final level of overrides during application invocation. For example, if the application was started with:

```bash
go run main.go -property-one "flag-value"
```

after flags are parsed, the above configuration struct would look like:

```go
Example{
  Property1: "flag-value",
  Property2: 2345,
  Property3: false,
  Property4: StructPropertyExample{
    SubProp1: "baz",
  },
}
```

### A note on environment variable name formation

Each property that has an `env` tag will be checked for a corresponding environment variable. When forming the name the of the environment variable to check for, configurator will take the value of that `env` tag, and prepend the value of `configurator.EnvPrefix` to it. For example, if the value of the `env` tag was `FOO`, and the value of `configurator.EnvPrefix` was `TEST_`, the environment variable configurator would search for would be `TEST_FOO`.

This means that `env` tag values should **not** be prefixed as well, unless you're planning to set `configurator.EnvPrefix` to an empty string.

**Also note:** environment variable names will always be upper case, even if the value of the `env` tag is lower case.

### A note on flag name formation

Flag names are formed based on a property's `env` tag, but do not contain the `configurator.EnvPrefix` and are lowercase. They also have any instances of `_` in the name replaced with `-`.
