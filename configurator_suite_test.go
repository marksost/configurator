// Test suite setup for the configurator package
package configurator

import (
	// Standard lib
	"io/ioutil"
	"testing"

	// Third-party
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

type (
	// Struct to use for testing configurator methods
	TestConfig struct {
		Foo         string            `default:"foo" json:"foo" env:"ENV_FOO"`
		FooEmpty    string            `default:"" json:"" env:""`
		Bar         int               `default:"1234" json:"bar" env:"ENV_BAR"`
		BarEmpty    int               `default:"" json:"" env:""`
		Baz         bool              `default:"true" json:"baz" env:"ENV_BAZ"`
		BazEmpty    bool              `default:""  json:"" env:""`
		Unsupported map[string]string `default:"doesnt-matter" json:"doesnt-matter" env:"DOESNT_MATTER"`
		Test        struct {
			Foo string `default:"test-foo" json:"test-foo" env:"ENV_TEST_FOO"`
		} `json:"test"`
	}
)

// Tests the configurator package
func TestConfigurator(t *testing.T) {
	// Register gomega fail handler
	RegisterFailHandler(Fail)

	// Have go's testing package run package specs
	RunSpecs(t, "Configurator Suite")
}

func init() {
	// Set logger output so as not to log during tests
	log.SetOutput(ioutil.Discard)
}
