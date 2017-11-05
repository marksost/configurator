package configurator

import (
	// Standard lib
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	// Third-party
	goutils "github.com/marksost/go-utils"
)

var (
	// EnvPrefix is a prefix used for environment variables
	// NOTE: This can be changed from outside this package before calling `InitializeConfig`
	EnvPrefix = "CONFIGURATOR_"
	// ConfigLocation is an environment variable with path to a config file
	// NOTE: This can be changed from outside this package before calling `InitializeConfig`
	ConfigLocation = EnvPrefix + "CONFIG"
)

// InitializeConfig is the main entrypoint to this package and takes in what is presumed to be
// a configuration struct with proper tags. It attempts to set up default values for each property,
// based on a `default` tag on the property. It then attempts to read in a configuration file
// based on a location retrieved from an environment variable (set with the above `ConfigLocation`) variable.
// This configuration file must be proper JSON and keys should map to `json` tags on the struct properties.
// NOTE: You may want to alter the value of that variable to be what your environment uses
// It will then attempt to read in environment variables to each struct property, using a concatenation
// of the `EnvPrefix` variable above and the value of an `env` tag for each property.
// Finally, it will parse command-line flags, using the unprefixed, lowercase version of the `env` tag value
// for each property
func InitializeConfig(c interface{}) {
	// Set up default values for configuration struct
	setDefaults(c)

	// Read in config file (if it exists) and set values on configuration struct
	setFromConfigFile(c)

	// Set environment variable-based values on configuration struct
	setFromEnvironment(c)

	// Parse command-line flags
	flag.Parse()
}

// setDefaults attempts to set default values for configuration properties
// based on a `default` tag assigned to each property
func setDefaults(c interface{}) {
	// Reflect value and pass to internal method
	handleDefaults(reflect.ValueOf(c))
}

// handleDefaults loops through a reflected value's fields based on their "kind",
// checks for a corresponding `default` tag and if found, sets it's value on the config
// NOTE: Abstracted from `setDefaults` to allow for struct recursion
func handleDefaults(v reflect.Value) {
	// Reflect indirectly to allow field looping
	val := reflect.Indirect(v)

	// Loop through fields
	for i := 0; i < val.NumField(); i++ {
		// Store field, kind, and tag value
		field := val.Field(i)
		kind := val.Field(i).Kind()
		tag := val.Type().Field(i).Tag.Get("default")

		// TO-DO: Logging?

		// Check for non-empty tag for non-struct types
		if tag == "" && kind != reflect.Struct {
			continue
		}

		// Handle field by it's "kind"
		switch kind {
		case reflect.Bool:
			field.SetBool(goutils.String2Bool(tag))
		case reflect.Int:
			field.SetInt(goutils.String2Int64(tag))
		case reflect.String:
			field.SetString(tag)
		case reflect.Struct:
			// Recurse
			handleDefaults(field.Addr())
		default:
			// TO-DO: Logging?
		}
	}
}

// setFromConfigFile attempts to unmarshal a configuration file's contents
// from JSON into the config struct, overriding any default values set previously
func setFromConfigFile(c interface{}) bool {
	var (
		contents []byte // Content of config file
		err      error  // Catch-all error
	)

	// Get config file contents
	if contents, err = getConfigFileContents(); err != nil {
		// TO-DO: Logging?
		return false
	}

	// Attempt to unmarshal JSON into config struct
	if err = json.Unmarshal(contents, &c); err != nil {
		// TO-DO: Logging?
		return false
	}

	return true
}

// getConfigFileContents attempts to load a JSON configuration file from disk and
// return it's contents if found, or an error if not
func getConfigFileContents() ([]byte, error) {
	var (
		contents []byte // Content of config file
		err      error  // Catch-all error
		file     string // Config file location, gotten from environment variable
	)

	// Allow for environment-level config file location override
	if file = os.Getenv(ConfigLocation); file == "" {
		return nil, fmt.Errorf("No valid file path detected under environment variable %s", ConfigLocation)
	}

	// TO-DO: Logging?

	// Attempt to load file
	if contents, err = ioutil.ReadFile(file); err != nil {
		// TO-DO: Logging?
		return nil, err
	}

	return contents, nil
}

// setFromEnvironment attempts to load environment variables matching
// the config struct's env tags, overriding any default or file-based values set previously
func setFromEnvironment(c interface{}) {
	// Reflect value and pass to internal method
	handleEnvironmentVariables(reflect.ValueOf(c))
}

// handleEnvironmentVariables loops through a reflected value's fields by their "kind",
// checks for a corresponding environment variable and if found, sets it
// both on the config and as a flag (when allowed)
// NOTE: Abstracted from `setFromEnvironment` to allow for struct recursion
func handleEnvironmentVariables(v reflect.Value) {
	// Reflect indirectly to allow field looping
	val := reflect.Indirect(v)

	// Loop through fields
	for i := 0; i < val.NumField(); i++ {
		// Store kind, env tag value, flag name, and OS value
		kind := val.Field(i).Kind()
		tag := EnvPrefix + val.Type().Field(i).Tag.Get("env")
		flagName := formFlagName(tag)
		// NOTE: Enforces upper-case env variables
		env := os.Getenv(strings.ToUpper(tag))

		// TO-DO: Logging?

		// Handle field by it's "kind"
		switch kind {
		case reflect.Bool:
			handleBoolEnvironmentVariable(val, i, flagName, env)
		case reflect.Int:
			handleIntEnvironmentVariable(val, i, flagName, env)
		case reflect.String:
			handleStringEnvironmentVariable(val, i, flagName, env)
		case reflect.Struct:
			// Recurse
			handleEnvironmentVariables(val.Field(i).Addr())
		default:
			// TO-DO: Logging?
		}
	}
}

// handleBoolEnvironmentVariable handles fields with a "kind" of bool
// Sets a field's value as well as a flag (when allowed)
func handleBoolEnvironmentVariable(v reflect.Value, i int, flagName string, env string) {
	// Store field
	field := v.Field(i)

	// Handle non-empty environment variable
	if env != "" {
		parsed, _ := strconv.ParseBool(env)
		field.SetBool(parsed)
	}

	// If allowed, set a flag
	// NOTE: Checks PkgPath for empty value, meaning the field is exported
	// and thus reflect's Interface method can return it's value
	// See https://golang.org/pkg/reflect/#StructField for more information
	if flag.Lookup(flagName) == nil && v.Type().Field(i).PkgPath == "" {
		ptr := field.Addr().Interface().(*bool)
		flag.BoolVar(ptr, flagName, field.Bool(), "")
	}
}

// handleIntEnvironmentVariable handles fields with a "kind" of int
// Sets a field's value as well as a flag (when allowed)
func handleIntEnvironmentVariable(v reflect.Value, i int, flagName string, env string) {
	// Store field
	field := v.Field(i)

	// Handle non-empty environment variable
	if env != "" {
		parsed, _ := strconv.ParseInt(env, 10, 0)
		field.SetInt(int64(parsed))
	}

	// If allowed, set a flag
	// NOTE: Checks PkgPath for empty value, meaning the field is exported
	// and thus reflect's Interface method can return it's value
	// See https://golang.org/pkg/reflect/#StructField for more information
	if flag.Lookup(flagName) == nil && v.Type().Field(i).PkgPath == "" {
		ptr := field.Addr().Interface().(*int)
		flag.IntVar(ptr, flagName, int(field.Int()), "")
	}
}

// handleStringEnvironmentVariable handles fields with a "kind" of string
// Sets a field's value as well as a flag (when allowed)
func handleStringEnvironmentVariable(v reflect.Value, i int, flagName string, env string) {
	// Store field
	field := v.Field(i)

	// Handle non-empty environment variable
	if env != "" {
		field.SetString(env)
	}

	// If allowed, set a flag
	// NOTE: Checks PkgPath for empty value, meaning the field is exported
	// and thus reflect's Interface method can return it's value
	// See https://golang.org/pkg/reflect/#StructField for more information
	if flag.Lookup(flagName) == nil && v.Type().Field(i).PkgPath == "" {
		ptr := field.Addr().Interface().(*string)
		flag.StringVar(ptr, flagName, field.String(), "")
	}
}

// formFlagName converts a field's tag corresponding to an environment variable
// into a string to use as a flag's name. Will strip the application prefix
// and replace underscores with hypens. Will also return the name in lowercase.
// Ex: IMG_FOO_BAR_BAZ => foo-bar-baz
func formFlagName(temp string) string {
	// Form flag name
	name := strings.TrimPrefix(strings.ToUpper(temp), EnvPrefix)
	name = strings.Replace(name, "_", "-", -1)

	return strings.ToLower(name)
}
