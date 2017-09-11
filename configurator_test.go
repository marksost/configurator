// Tests the configurator.go file
package configurator

import (
	// Standard lib
	"flag"
	"os"
	"path"
	"reflect"

	// Third-party
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("configurator.go", func() {
	var (
		// Mock config to use throughout tests
		testConfig *TestConfig
	)

	BeforeEach(func() {
		// Set up test config
		testConfig = &TestConfig{}
	})

	Describe("`InitializeConfig` method", func() {
		BeforeEach(func() {
			// Set config env var
			os.Setenv(ConfigLocation, path.Join("test/data/valid-config.json"))

			// Unset test environment variables
			os.Unsetenv(EnvPrefix + "ENV_FOO")
			os.Unsetenv(EnvPrefix + "ENV_BAR")
			os.Unsetenv(EnvPrefix + "ENV_BAZ")
		})

		Context("Without config file or environment variables set", func() {
			BeforeEach(func() {
				// Ensure config env var is unset
				os.Unsetenv(ConfigLocation)
			})

			It("Uses default values", func() {
				// Call method
				InitializeConfig(testConfig)

				// Verify values were set
				Expect(testConfig.Foo).To(Equal("foo"))
				Expect(testConfig.Bar).To(Equal(1234))
				Expect(testConfig.Baz).To(BeTrue())
				Expect(testConfig.Test.Foo).To(Equal("test-foo"))
			})
		})

		Context("With config file available but no environment variables set", func() {
			It("Uses default values and config file overrides", func() {
				// Call method
				InitializeConfig(testConfig)

				// Verify values were set
				Expect(testConfig.Foo).To(Equal("abcd"))
				Expect(testConfig.Bar).To(Equal(1234))
				Expect(testConfig.Baz).To(BeTrue())
				Expect(testConfig.Test.Foo).To(Equal("bcde"))
			})
		})

		Context("With environment variables set", func() {
			BeforeEach(func() {
				// Set test environment variables
				os.Setenv(EnvPrefix+"ENV_FOO", "foo")
				os.Setenv(EnvPrefix+"ENV_BAR", "1234")
				os.Setenv(EnvPrefix+"ENV_BAZ", "0")
			})

			It("Uses default values, config file overrides, and environment variable overrides", func() {
				// Call method
				InitializeConfig(testConfig)

				// Verify values were set
				Expect(testConfig.Foo).To(Equal("foo"))
				Expect(testConfig.Bar).To(Equal(1234))
				Expect(testConfig.Baz).To(BeFalse())
				Expect(testConfig.Test.Foo).To(Equal("bcde"))
			})
		})
	})

	Describe("Default handling methods", func() {
		Describe("`handleDefaults` method", func() {
			It("Sets configuration values based on default values and types", func() {
				// Call method
				handleDefaults(reflect.ValueOf(testConfig))

				// Verify values were set
				Expect(testConfig.Foo).To(Equal("foo"))
				Expect(testConfig.Bar).To(Equal(1234))
				Expect(testConfig.Baz).To(BeTrue())
				Expect(testConfig.Test.Foo).To(Equal("test-foo"))
			})
		})
	})

	Describe("Config file handling methods", func() {
		Describe("`setFromConfigFile` method", func() {
			Context("When no config location environment variable is set", func() {
				BeforeEach(func() {
					// Ensure config env var is unset
					os.Unsetenv(ConfigLocation)
				})

				It("Returns false", func() {
					// Verify return value
					Expect(setFromConfigFile(testConfig)).To(BeFalse())
				})
			})

			Context("When the configuration file doesn't exist", func() {
				BeforeEach(func() {
					// Set config env var
					os.Setenv(ConfigLocation, path.Join("test/data/doesnt-exist.json"))
				})

				It("Returns false", func() {
					// Verify return value
					Expect(setFromConfigFile(testConfig)).To(BeFalse())
				})
			})

			Context("When the configuration file contains invalid JSON", func() {
				BeforeEach(func() {
					// Set config env var
					os.Setenv(ConfigLocation, path.Join("test/data/invalid-config.json"))
				})

				It("Returns false", func() {
					// Verify return value
					Expect(setFromConfigFile(testConfig)).To(BeFalse())
				})
			})

			Context("When the configuration file contains valid JSON", func() {
				BeforeEach(func() {
					// Set config env var
					os.Setenv(ConfigLocation, path.Join("test/data/valid-config.json"))
				})

				It("Reads in the configuration, sets values, and returns true", func() {
					// Verify return value
					Expect(setFromConfigFile(testConfig)).To(BeTrue())

					// Verify values were set
					Expect(testConfig.Foo).To(Equal("abcd"))
					Expect(testConfig.Test.Foo).To(Equal("bcde"))
				})
			})
		})
	})

	Describe("Environment variable handling methods", func() {
		BeforeEach(func() {
			// Set test environment variables
			os.Setenv(EnvPrefix+"ENV_FOO", "foo")
			os.Setenv(EnvPrefix+"ENV_BAR", "1234")
			os.Setenv(EnvPrefix+"ENV_BAZ", "1")
			os.Setenv(EnvPrefix+"ENV_TEST_FOO", "test-foo")
		})

		Describe("`handleEnvironmentVariables` method", func() {
			It("Sets configuration values based on environment variable values and types", func() {
				// Call method
				handleEnvironmentVariables(reflect.ValueOf(testConfig))

				// Verify values were set
				Expect(testConfig.Foo).To(Equal("foo"))
				Expect(testConfig.Bar).To(Equal(1234))
				Expect(testConfig.Baz).To(BeTrue())
				Expect(testConfig.Test.Foo).To(Equal("test-foo"))
			})

			It("Sets flags based on environment variables that set set", func() {
				// Call method
				handleEnvironmentVariables(reflect.ValueOf(testConfig))

				// Verify flags were set
				Expect(flag.Lookup("env-foo")).To(Not(BeNil()))
				Expect(flag.Lookup("env-bar")).To(Not(BeNil()))
				Expect(flag.Lookup("env-baz")).To(Not(BeNil()))
				Expect(flag.Lookup("env-test-foo")).To(Not(BeNil()))
			})
		})

		Describe("`formFlagName` method", func() {
			var (
				// Input for `formFlagName` input
				input map[string]string
			)

			BeforeEach(func() {
				// Set input
				input = map[string]string{
					"test":                "test",
					"TEST":                "test",
					"test_foo":            "test-foo",
					EnvPrefix + "foo_bar": "foo-bar",
					EnvPrefix + "FOO_BAR": "foo-bar",
				}
			})

			It("Returns a formatted flag name", func() {
				// Loop through test data
				for input, expected := range input {
					// Call method
					actual := formFlagName(input)

					// Verify result
					Expect(actual).To(Equal(expected))
				}
			})
		})
	})
})
