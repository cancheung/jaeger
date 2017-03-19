package app

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/zap"
	"github.com/uber/jaeger-lib/metrics"
	"github.com/uber/jaeger/thrift-gen/jaeger"
	"github.com/uber/jaeger/thrift-gen/zipkincore"
	"gopkg.in/yaml.v2"

	"code.uber.internal/infra/jaeger/oss/pkg/discovery"
)

func TestConfigFile(t *testing.T) {
	cfg := Builder{}
	data, err := ioutil.ReadFile("testdata/test_config.yaml")
	require.NoError(t, err)
	err = yaml.Unmarshal(data, &cfg)
	require.NoError(t, err)
	assert.Len(t, cfg.Processors, 3)
	for i := range cfg.Processors {
		cfg.Processors[i].applyDefaults()
		cfg.Processors[i].Server.applyDefaults()
	}
	assert.Equal(t, ProcessorConfiguration{
		Model:    zipkinModel,
		Protocol: compactProtocol,
		Workers:  10,
		Server: ServerConfiguration{
			QueueSize:     1000,
			MaxPacketSize: 65000,
			HostPort:      "1.1.1.1:5775",
		},
	}, cfg.Processors[0])
	assert.Equal(t, ProcessorConfiguration{
		Model:    jaegerModel,
		Protocol: compactProtocol,
		Workers:  10,
		Server: ServerConfiguration{
			QueueSize:     1000,
			MaxPacketSize: 65000,
			HostPort:      "2.2.2.2:6831",
		},
	}, cfg.Processors[1])
	assert.Equal(t, ProcessorConfiguration{
		Model:    jaegerModel,
		Protocol: binaryProtocol,
		Workers:  20,
		Server: ServerConfiguration{
			QueueSize:     2000,
			MaxPacketSize: 65001,
			HostPort:      "3.3.3.3:6832",
		},
	}, cfg.Processors[2])
	assert.Equal(t, "4.4.4.4:5778", cfg.SamplingServer.HostPort)
}

func TestConfigWithDiscovery(t *testing.T) {
	cfg := &Builder{}
	discoverer := discovery.FixedDiscoverer([]string{"1.1.1.1:80"})
	cfg.WithDiscoverer(discoverer)
	_, err := cfg.CreateAgent(metrics.NullFactory, zap.New(zap.NullEncoder()))
	assert.EqualError(t, err, "cannot enable service discovery: both discovery.Discoverer and discovery.Notifier must be specified")

	cfg = &Builder{}
	notifier := &discovery.Dispatcher{}
	cfg.WithDiscoverer(discoverer).WithDiscoveryNotifier(notifier)
	agent, err := cfg.CreateAgent(metrics.NullFactory, zap.New(zap.NullEncoder()))
	assert.NoError(t, err)
	assert.NotNil(t, agent)
}

func TestConfigWithSingleCollector(t *testing.T) {
	cfg := &Builder{
		CollectorHostPort: "127.0.0.1:9876",
	}
	agent, err := cfg.CreateAgent(metrics.NullFactory, zap.New(zap.NullEncoder()))
	assert.NoError(t, err)
	assert.NotNil(t, agent)
}

type fakeReporter struct{}

func (fr fakeReporter) EmitZipkinBatch(spans []*zipkincore.Span) (err error) {
	return nil
}

func (fr fakeReporter) EmitBatch(batch *jaeger.Batch) (err error) {
	return nil
}

func TestConfigWithExtraReporter(t *testing.T) {
	cfg := &Builder{}
	cfg.WithReporter(fakeReporter{})
	agent, err := cfg.CreateAgent(metrics.NullFactory, zap.New(zap.NullEncoder()))
	assert.NoError(t, err)
	assert.NotNil(t, agent)
}

func TestConfigWithProcessorErrors(t *testing.T) {
	testCases := []struct {
		model       model
		protocol    protocol
		hostPort    string
		err         string
		errContains string
	}{
		{protocol: protocol("bad"), err: "cannot find protocol factory for protocol bad"},
		{protocol: compactProtocol, model: model("bad"), err: "cannot find agent processor for data model bad"},
		{protocol: compactProtocol, model: jaegerModel, err: "no host:port provided for udp server: {QueueSize:1000 MaxPacketSize:65000 HostPort:}"},
		{protocol: compactProtocol, model: zipkinModel, hostPort: "bad-host-port", errContains: "bad-host-port"},
	}
	for _, tc := range testCases {
		testCase := tc // capture loop var
		cfg := &Builder{
			Processors: []ProcessorConfiguration{
				{
					Model:    testCase.model,
					Protocol: testCase.protocol,
					Server: ServerConfiguration{
						HostPort: testCase.hostPort,
					},
				},
			},
		}
		_, err := cfg.CreateAgent(metrics.NullFactory, zap.New(zap.NullEncoder()))
		assert.Error(t, err)
		if testCase.err != "" {
			assert.EqualError(t, err, testCase.err)
		} else if testCase.errContains != "" {
			assert.True(t, strings.Contains(err.Error(), testCase.errContains), "error must contain %s", testCase.errContains)
		}
	}
}
