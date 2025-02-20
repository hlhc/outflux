//go:build integration

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/timescale/outflux/internal/cli"
	"github.com/timescale/outflux/internal/schemamanagement/schemaconfig"
	"github.com/timescale/outflux/internal/testutils"
)

func TestSchemaTransfer(t *testing.T) {
	db := "test_schema_transfer"
	field := "field1"
	value := 1
	tags := make(map[string]string)
	fieldValues := make(map[string]interface{})
	fieldValues[field] = value
	measures := []string{"test", "test 1", "test-2"}

	if err := testutils.PrepareServersForITest(db); err != nil {
		t.Fatalf("could not prepare servers: %v", err)
	}

	defer testutils.ClearServersAfterITest(db)
	appContext := initAppContext()

	dbConn, err := testutils.OpenTSConn(db)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()
	for _, measure := range measures {
		err := testutils.CreateInfluxMeasure(db, measure, []*map[string]string{&tags}, []*map[string]interface{}{&fieldValues})
		if err != nil {
			t.Fatalf("could not create measure: %v", err)
		}

		connConf := &cli.ConnectionConfig{
			InputHost:          testutils.InfluxHost,
			InputDb:            db,
			InputMeasures:      []string{measure},
			OutputDbConnString: fmt.Sprintf(testutils.TsConnStringTemplate, db),
		}
		config := &cli.MigrationConfig{
			ChunkSize:            1,
			OutputSchemaStrategy: schemaconfig.DropAndCreate,
			SchemaOnly:           true,
		}

		err = transferSchema(appContext, connConf, config)
		if err != nil {
			t.Fatal(err)
		}

		rows, err := dbConn.Query(fmt.Sprintf(`SELECT count(*) FROM  "%s"`, measure))
		if err != nil {
			t.Fatal(err)
		}
		var count int
		if !rows.Next() {
			t.Fatal("couldn't check state of TS DB")
		}

		err = rows.Scan(&count)
		if err != nil {
			t.Fatal("couldn't check state of TS DB")
		}

		if count != 0 {
			t.Errorf("expected no rows in the output database, %d found", count)
		}
		rows.Close()
	}
}

func TestOutputConnOverridesEnvVars(t *testing.T) {
	// Set up servers
	db := "test_output_con_overrides"
	measure := "test"
	field := "field1"
	value := 1
	tags := make(map[string]string)
	fieldValues := make(map[string]interface{})
	fieldValues[field] = value
	if err := testutils.PrepareServersForITest(db); err != nil {
		t.Fatalf("could not prepare servers: %v", err)
	}

	err := testutils.CreateInfluxMeasure(db, measure, []*map[string]string{&tags}, []*map[string]interface{}{&fieldValues})
	if err != nil {
		t.Fatalf("could not create influx measure: %v", err)
	}

	defer testutils.ClearServersAfterITest(db)

	// Three PG environment variables determining database and password
	os.Setenv("PGDATABASE", "wrong_db")
	os.Setenv("PGPORT", "5433")
	os.Setenv("PGPASSWORD", "postgres")

	connConf := &cli.ConnectionConfig{
		InputHost:     testutils.InfluxHost,
		InputDb:       db,
		InputMeasures: []string{measure},
	}
	config := &cli.MigrationConfig{
		ChunkSize:            1,
		OutputSchemaStrategy: schemaconfig.DropAndCreate,
		SchemaOnly:           true,
	}
	appContext := initAppContext()

	// connection should fail, wrong db
	err = transferSchema(appContext, connConf, config)
	if err == nil {
		t.Fatal("expected error, none received")
	}

	// Conn String that will override database and user
	connString := fmt.Sprintf("user=postgres dbname=%s", db)
	connConf.OutputDbConnString = connString
	err = transferSchema(appContext, connConf, config)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
