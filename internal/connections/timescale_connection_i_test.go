//go:build integration
// +build integration

package connections

import (
	"os"
	"testing"

	"github.com/timescale/outflux/internal/testutils"
)

func TestNewConnection(t *testing.T) {
	db := "test_new_conn"
	if err := testutils.CreateTimescaleDb(db); err != nil {
		t.Fatalf("could not prepare db: %v", err)
	}

	defer testutils.DeleteTimescaleDb(db)

	goodEnv := map[string]string{
		"PGPORT":     "5433",
		"PGUSER":     "postgres",
		"PGPASSWORD": "postgres",
		"PGDATABASE": db,
	}

	badEnv := map[string]string{
		"PGPORT":     "5433",
		"PGUSER":     "postgres",
		"PGPASSWORD": "postgres",
		"PGDATABASE": "wrong_db",
	}
	connService := &defaultTSConnectionService{}
	testCases := []struct {
		desc      string
		conn      string
		env       map[string]string
		expectErr bool
	}{
		{desc: "nothing is set, env is empty", expectErr: true},
		{desc: "environment is set, no overrides", env: goodEnv},
		{desc: "environment is set, overrides make is bad", env: goodEnv, conn: "dbname=wrong_db", expectErr: true},
		{desc: "environment is set badly, overrides make it good", env: badEnv, conn: "dbname=" + db},
	}

	for _, tc := range testCases {
		// make sure the environment is only that in tc.env
		os.Clearenv()
		for k, v := range tc.env {
			os.Setenv(k, v)
		}
		res, err := connService.NewConnection(tc.conn)
		if err != nil && !tc.expectErr {
			t.Fatalf("%s\nunexpected error: %v", tc.desc, err)
		} else if err == nil && tc.expectErr {
			res.Close()
			t.Fatalf("%s\nexpected error, none received", tc.desc)
		}

		if tc.expectErr {
			continue
		}

		rows, err := res.Query("SELECT 1")
		if err != nil {
			t.Fatalf("could execute query with established connection")
		}

		if !rows.Next() {
			t.Fatalf("no result returned for SELECT 1")
		} else {
			var dest int
			rows.Scan(&dest)
			if dest != 1 {
				t.Fatalf("expected 1, got %d", dest)
			}
		}

		rows.Close()
		res.Close()
	}
}
