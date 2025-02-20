package influx

import (
	"fmt"
	"testing"

	"github.com/timescale/outflux/internal/idrf"

	influx "github.com/influxdata/influxdb/client/v2"
)

func TestNewDataSetConstructor(t *testing.T) {
	newDataSetConstructor("", "rp", true, nil, nil, nil)
}

func TestConstruct(t *testing.T) {
	genError := fmt.Errorf("generic error")
	tags := []*idrf.Column{{Name: "tag", DataType: idrf.IDRFString}}
	fields := []*idrf.Column{{Name: "field", DataType: idrf.IDRFBoolean}}
	testCases := []struct {
		desc        string
		tags        []*idrf.Column
		tagsErr     error
		fields      []*idrf.Column
		fieldsErr   error
		expectedErr bool
	}{
		{
			desc:        "Error on discover tags",
			tagsErr:     genError,
			expectedErr: true,
		}, {
			desc:        "Error on discover fields",
			tags:        tags,
			fieldsErr:   genError,
			expectedErr: true,
		}, {
			desc:   "All good",
			tags:   tags,
			fields: fields,
		},
	}

	for _, tc := range testCases {
		mock := &mocker{tags: tc.tags, tagsErr: tc.tagsErr, fields: tc.fields, fieldsErr: tc.fieldsErr}
		constructor := defaultDSConstructor{
			tagExplorer:   mock,
			fieldExplorer: mock,
		}

		res, err := constructor.construct("a")
		if err != nil && !tc.expectedErr {
			t.Errorf("unexpected error %v", err)
		} else if err == nil && tc.expectedErr {
			t.Errorf("expected error, none received")
		}

		if tc.expectedErr {
			continue
		}

		if res.DataSetName != "a" {
			t.Errorf("expected data set to be named: a, got: %s", res.DataSetName)
		}

		if len(res.Columns) != 1+len(tags)+len(fields) { // time, tags, fields
			t.Errorf("exected %d columns, got %d", 1+len(tags)+len(fields), len(res.Columns))
		}

		if res.TimeColumn != res.Columns[0].Name {
			t.Errorf("expectd time column to be first in columns array")
		}
	}
}

type mocker struct {
	tags      []*idrf.Column
	tagsErr   error
	fields    []*idrf.Column
	fieldsErr error
}

func (m *mocker) DiscoverMeasurementTags(influxClient influx.Client, db, rp, measure string) ([]*idrf.Column, error) {
	return m.tags, m.tagsErr
}

func (m *mocker) DiscoverMeasurementFields(influxClient influx.Client, db, rp, measurement string, convertIntToFloat bool) ([]*idrf.Column, error) {
	return m.fields, m.fieldsErr
}
