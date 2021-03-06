package config

import (
	"testing"
	"time"

	"github.com/Shopify/sarama"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type ConfigSuite struct{}

var _ = Suite(&ConfigSuite{})

func (s *ConfigSuite) TestFromYAMLNoProxies(c *C) {
	// When
	_, err := FromYAML([]byte(""))

	// Then
	c.Assert(err.Error(), Equals, "invalid config parameter: "+
		"at least on proxy must be configured")
}

// Configuration fields that are not explicitly mentioned if the YAML data are
// left intact.
func (s *ConfigSuite) TestFromYAML(c *C) {
	data := []byte("" +
		"proxies:\n" +
		"  bar:\n" +
		"    client_id: foo\n" +
		"    kafka:\n" +
		"      seed_peers:\n" +
		"        - 192.168.19.2:9092\n" +
		"        - 192.168.19.3:9092\n" +
		"    consumer:\n" +
		"      long_polling_timeout: 5s\n")

	// When
	appCfg, err := FromYAML(data)

	// Then
	c.Assert(err, IsNil)

	expected := DefaultApp("bar")
	expected.Proxies["bar"].ClientID = "foo"
	expected.Proxies["bar"].Kafka.SeedPeers = []string{"192.168.19.2:9092", "192.168.19.3:9092"}
	expected.Proxies["bar"].Consumer.LongPollingTimeout = 5 * time.Second
	c.Assert(appCfg, DeepEquals, expected)
}

// If YAML data is invalid then the original config is not changed.
func (s *ConfigSuite) TestFromYAMLInvalid(c *C) {
	data := []byte("" +
		"proxies:\n" +
		"  default:\n" +
		"    client_id: foo\n" +
		"    kafka:\n" +
		"      seed_peers:\n" +
		"        - 192.168.19.2:9092\n" +
		"        - 192.168.19.3:9092\n" +
		"    consumer:\n" +
		"      long_polling_timeout: Kaboom!\n")

	// When
	_, err := FromYAML(data)

	// Then
	c.Assert(err.Error(), Equals, "failed to parse config: "+
		"yaml: unmarshal errors:\n"+
		"  line 9: cannot unmarshal !!str `Kaboom!` into time.Duration")
}

// The first proxy mentioned is returned as default.
func (s *ConfigSuite) TestFromYAMLDefault(c *C) {
	data := []byte("" +
		"proxies:\n" +
		"  foo:\n" +
		"    client_id: foo_id\n" +
		"  bar:\n" +
		"    client_id: bar_id\n" +
		"  bazz:\n" +
		"    client_id: bazz_id\n")

	// When
	appCfg, err := FromYAML(data)

	// Then
	c.Assert(err, IsNil)
	c.Assert(appCfg.DefaultCluster, Equals, "foo")
	c.Assert(appCfg.Proxies["foo"].ClientID, Equals, "foo_id")
	c.Assert(appCfg.Proxies["bar"].ClientID, Equals, "bar_id")
	c.Assert(appCfg.Proxies["bazz"].ClientID, Equals, "bazz_id")
}

// default.yaml contains the same configuration as returned by Default()
func (s *ConfigSuite) TestFromYAMLFile(c *C) {
	// When
	appCfg, err := FromYAMLFile("../default.yaml")

	// Then
	c.Assert(err, IsNil)
	expected := DefaultApp("default")
	expected.Proxies["default"].ClientID = "ID"
	expected.Proxies["default"].Kafka.Version.Set(sarama.V0_10_2_1)
	expected.Logging = append(expected.Logging, LoggerCfg{
		Name:     "console",
		Severity: "info",
	})
	appCfg.Proxies["default"].ClientID = "ID"
	c.Assert(appCfg, DeepEquals, expected)
}

func (s *ConfigSuite) TestFromYAMLTLS(c *C) {
	// When
	appCfg, err := FromYAMLFile("../testdata/tls.yaml")

	// Then
	c.Assert(err, IsNil)
	c.Assert(appCfg.TLS.CertPath, Equals, "/usr/local/etc/server.crt")
	c.Assert(appCfg.TLS.KeyPath, Equals, "/usr/local/etc/server.key")
}

func (s *ConfigSuite) TestFromYAMLCustomAddresses(c *C) {
	// When
	appCfg, err := FromYAMLFile("../testdata/custom-hostname.yaml")

	// Then
	c.Assert(err, IsNil)

	expected := DefaultApp("default")
	expected.TCPAddr = "foo.bar:443"
	expected.GRPCAddr = "bar.baz:50000"
	expected.UnixAddr = "/var/run/kafka-pixy.sock"
	c.Assert(appCfg.TCPAddr, Equals, expected.TCPAddr)
	c.Assert(appCfg.GRPCAddr, Equals, expected.GRPCAddr)
	c.Assert(appCfg.UnixAddr, Equals, expected.UnixAddr)
}

func (s *ConfigSuite) TestFromYAMLKafkaTLS(c *C) {
	// When
	appCfg, err := FromYAMLFile("../testdata/kafka-tls.yaml")

	// Then
	c.Assert(err, IsNil)

	kafkaCfg := appCfg.Proxies["default"].Kafka
	c.Assert(kafkaCfg.TLSEnabled, Equals, true)
	c.Assert(kafkaCfg.CACertFile, Equals, "../testdata/ca.crt")
	c.Assert(kafkaCfg.ClientCertFile, Equals, "../testdata/client.crt")
	c.Assert(kafkaCfg.ClientCertKeyFile, Equals, "../testdata/client.key")
	c.Assert(kafkaCfg.InsecureSkipVerify, Equals, false)
}
