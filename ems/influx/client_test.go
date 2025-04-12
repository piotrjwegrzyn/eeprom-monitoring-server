package influx

import (
	"fmt"
	"testing"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_InsertMeasurements(t *testing.T) {
	mock := &influxMock{}
	testHostname := "test-hostname"

	client := Client{
		config: Config{
			Bucket: "test-bucket",
			Org:    "test-org",
		},
		influxClient: mock,
	}

	client.InsertMeasurements(testHostname, "test-interface", Measurement{Temperature: 123})

	require.NotNil(t, mock.point)
	assert.Contains(t, fmt.Sprintf("%v", *mock.point), testHostname)
}

type influxMock struct {
	called   int
	point    *write.Point
	writeAPI *writeAPIMock
}

func (i *influxMock) WriteAPI(org string, bucket string) api.WriteAPI {
	i.called++
	i.writeAPI = &writeAPIMock{point: &i.point}
	return i.writeAPI
}

type writeAPIMock struct {
	point **write.Point
}

func (w writeAPIMock) WriteRecord(string)                             {}
func (w writeAPIMock) WritePoint(p *write.Point)                      { *w.point = p }
func (w writeAPIMock) Flush()                                         {}
func (w writeAPIMock) Errors() <-chan error                           { return make(<-chan error) }
func (w writeAPIMock) SetWriteFailedCallback(api.WriteFailedCallback) {}
