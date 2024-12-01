//go:build integration

package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	gocmp "github.com/google/go-cmp/cmp"
	"github.com/kelseyhightower/envconfig"
	"github.com/pressly/goose/v3"
)

//go:embed sqlc/migrations/*.sql
var migrations embed.FS

const (
	dbVersionTable = "db_version"
	migrationsDir  = "sqlc/migrations"
)

func TestMain(m *testing.M) {
	conn, err := connect()
	if err != nil {
		fmt.Printf("unable to connect to database: %v", err)
		os.Exit(1)
	}

	goose.SetBaseFS(migrations)

	if err = goose.SetDialect("mysql"); err != nil {
		fmt.Printf("unable to set dialect: %v", err)
		os.Exit(1)
	}

	goose.SetTableName(dbVersionTable)

	if _, err = goose.EnsureDBVersion(conn); err != nil {
		fmt.Printf("unable to ensure db version: %v", err)
		os.Exit(1)
	}

	if err = goose.Reset(conn, migrationsDir); err != nil {
		fmt.Printf("unable to reset migrations: %v", err)
		os.Exit(1)
	}

	if err = goose.Up(conn, migrationsDir); err != nil {
		fmt.Printf("unable to run migrations: %v", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	if err = goose.Reset(conn, migrationsDir); err != nil {
		fmt.Printf("unable to reset migrations: %v", err)
		os.Exit(1)
	}

	if _, err := conn.Exec("DROP TABLE " + dbVersionTable); err != nil {
		fmt.Printf("unable to drop table: %v", err)
		os.Exit(1)
	}

	os.Exit(exitCode)
}

func cleanup(tables ...string) func(*testing.T, *sql.DB) {
	return func(t *testing.T, db *sql.DB) {
		t.Helper()

		for _, table := range tables {
			if _, err := db.Exec("DELETE FROM " + table); err != nil {
				t.Fatalf("unable to delete from table: %v", err)
			}
		}
	}
}

func exec(q string) func(*testing.T, *sql.DB) {
	return func(t *testing.T, db *sql.DB) {
		t.Helper()

		if _, err := db.Exec(q); err != nil {
			t.Fatalf("unable to execute query: %v", err)
		}
	}
}

func count(table string) func(*testing.T, *sql.DB) int {
	return func(t *testing.T, db *sql.DB) int {
		t.Helper()

		var r int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&r); err != nil {
			t.Fatalf("unable to count elements: %v", err)
		}

		return r
	}
}

func ptr[T any](v T) *T {
	return &v
}

func connect() (*sql.DB, error) {
	var cfg struct {
		Name     string `envconfig:"DB_NAME" required:"true" validate:"required"`
		User     string `envconfig:"DB_USER" required:"true" validate:"required"`
		Password string `envconfig:"DB_PASSWORD" required:"true" validate:"required"`
		Host     string `envconfig:"DB_HOST" required:"true" validate:"required"`
		Port     string `envconfig:"DB_PORT" required:"true" validate:"required"`
	}

	if err := envconfig.Process("EMS", &cfg); err != nil {
		return nil, err
	}

	return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	))
}

func TestDB_CreateDevice(t *testing.T) {
	type args struct {
		ctx    context.Context
		device Device
	}
	type want struct {
		count *int
		err   error
	}
	type database struct {
		prepare func(*testing.T, *sql.DB)
		cleanup func(*testing.T, *sql.DB)
	}
	tests := []struct {
		name     string
		args     args
		want     want
		database database
	}{
		{
			name: "creates device",
			args: args{
				ctx: context.Background(),
				device: Device{
					Hostname:  "hostname2",
					IPAddress: "10.0.0.2",
					Login:     "user2",
				},
			},
			want: want{
				count: ptr(2),
			},
			database: database{
				prepare: exec(`INSERT INTO devices(id, hostname, ip, login, connected)
VALUES (1,'hostname1','10.0.0.1','user1','2024-05-22 00:00:00');`),
				cleanup: cleanup("devices"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := connect()
			if err != nil {
				t.Fatalf("unable to connect to database: %v", err)
			}

			if tc.database.prepare != nil {
				tc.database.prepare(t, conn)
			}
			if tc.database.cleanup != nil {
				t.Cleanup(func() { tc.database.cleanup(t, conn) })
			}

			db := New(conn)
			err = db.CreateDevice(tc.args.ctx, tc.args.device)

			errComp := gocmp.Comparer(func(x, y error) bool {
				return x.Error() == y.Error()
			})

			if diff := gocmp.Diff(err, tc.want.err, errComp); diff != "" {
				t.Errorf("error mismatch (-got +want):\n%s", diff)
			}

			if tc.want.count != nil {
				got := count("devices")(t, conn)
				if diff := gocmp.Diff(got, *tc.want.count); diff != "" {
					t.Errorf("devices count mismatch (-got +want):\n%s", diff)
				}
			}
		})
	}
}

func TestDB_Device(t *testing.T) {
	type args struct {
		ctx context.Context
		id  uint
	}
	type want struct {
		device Device
		err    error
	}
	type database struct {
		prepare func(*testing.T, *sql.DB)
		cleanup func(*testing.T, *sql.DB)
	}
	tests := []struct {
		name     string
		args     args
		want     want
		database database
	}{
		{
			name: "returns device wih specific ID",
			args: args{
				ctx: context.Background(),
				id:  uint(2),
			},
			want: want{
				device: Device{
					ID:         2,
					Hostname:   "hostname2",
					IPAddress:  "10.0.0.2",
					Login:      "user2",
					Keyfile:    []byte{},
					Connected:  time.Date(2024, 5, 22, 0, 0, 0, 0, time.UTC),
					LastStatus: -1,
				},
			},
			database: database{
				prepare: exec(`INSERT INTO devices(id, hostname, ip, login, connected)
VALUES (1,'hostname1','10.0.0.1','user1','2024-05-22 00:00:00'),
       (2,'hostname2','10.0.0.2','user2','2024-05-22 00:00:00'),
	   (3,'hostname3','10.0.0.3','user3','2024-05-22 00:00:00');`),
				cleanup: cleanup("devices"),
			},
		},
		{
			name: "returns errors with no device",
			args: args{
				ctx: context.Background(),
				id:  uint(2),
			},
			want: want{
				err: errors.New("sql: no rows in result set"),
			},
			database: database{
				cleanup: cleanup("devices"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := connect()
			if err != nil {
				t.Fatalf("unable to connect to database: %v", err)
			}

			if tc.database.prepare != nil {
				tc.database.prepare(t, conn)
			}
			if tc.database.cleanup != nil {
				t.Cleanup(func() { tc.database.cleanup(t, conn) })
			}

			db := New(conn)

			device, err := db.Device(tc.args.ctx, tc.args.id)

			errComp := gocmp.Comparer(func(x, y error) bool {
				return x.Error() == y.Error()
			})

			if diff := gocmp.Diff(err, tc.want.err, errComp); diff != "" {
				t.Errorf("error mismatch (-got +want):\n%s", diff)
			}

			deviceComp := gocmp.Comparer(func(x, y Device) bool {
				return x.ID == y.ID &&
					x.Hostname == y.Hostname &&
					x.IPAddress == y.IPAddress &&
					x.Login == y.Login &&
					x.Password == y.Password &&
					string(x.Keyfile) == string(y.Keyfile) &&
					x.Connected == y.Connected &&
					x.LastStatus == y.LastStatus
			})

			if diff := gocmp.Diff(device, tc.want.device, deviceComp); diff != "" {
				t.Errorf("device read mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestDB_Devices(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type want struct {
		devices []Device
		err     error
	}
	type database struct {
		prepare func(*testing.T, *sql.DB)
		cleanup func(*testing.T, *sql.DB)
	}
	tests := []struct {
		name     string
		args     args
		want     want
		database database
	}{
		{
			name: "returns all devices",
			args: args{
				ctx: context.Background(),
			},
			want: want{
				devices: []Device{
					{
						ID:         1,
						Hostname:   "hostname1",
						IPAddress:  "10.0.0.1",
						Login:      "user1",
						Keyfile:    []byte{},
						Connected:  time.Date(2024, 5, 22, 0, 0, 0, 0, time.UTC),
						LastStatus: -1,
					}, {
						ID:         2,
						Hostname:   "hostname2",
						IPAddress:  "10.0.0.2",
						Login:      "user2",
						Keyfile:    []byte{},
						Connected:  time.Date(2024, 5, 22, 0, 0, 0, 0, time.UTC),
						LastStatus: -1,
					},
				},
			},
			database: database{
				prepare: exec(`INSERT INTO devices(id, hostname, ip, login, connected)
VALUES (1,'hostname1','10.0.0.1','user1','2024-05-22 00:00:00'),
       (2,'hostname2','10.0.0.2','user2','2024-05-22 00:00:00');`),
				cleanup: cleanup("devices"),
			},
		},
		{
			name: "returns no error with no devices",
			args: args{
				ctx: context.Background(),
			},
			want: want{
				devices: []Device{},
				err:     nil,
			},
			database: database{
				cleanup: cleanup("devices"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := connect()
			if err != nil {
				t.Fatalf("unable to connect to database: %v", err)
			}

			if tc.database.prepare != nil {
				tc.database.prepare(t, conn)
			}
			if tc.database.cleanup != nil {
				t.Cleanup(func() { tc.database.cleanup(t, conn) })
			}

			db := New(conn)

			devices, err := db.Devices(tc.args.ctx)

			errComp := gocmp.Comparer(func(x, y error) bool {
				return x.Error() == y.Error()
			})

			if diff := gocmp.Diff(err, tc.want.err, errComp); diff != "" {
				t.Errorf("error mismatch (-got +want):\n%s", diff)
			}

			deviceComp := gocmp.Comparer(func(x, y Device) bool {
				return x.ID == y.ID &&
					x.Hostname == y.Hostname &&
					x.IPAddress == y.IPAddress &&
					x.Login == y.Login &&
					x.Password == y.Password &&
					string(x.Keyfile) == string(y.Keyfile) &&
					x.Connected == y.Connected &&
					x.LastStatus == y.LastStatus
			})

			if diff := gocmp.Diff(devices, tc.want.devices, deviceComp); diff != "" {
				t.Errorf("devices read mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestDB_UpdateDevice(t *testing.T) {
	type args struct {
		ctx    context.Context
		device Device
	}
	type want struct {
		devices []Device
		err     error
	}
	type database struct {
		prepare func(*testing.T, *sql.DB)
		cleanup func(*testing.T, *sql.DB)
	}
	tests := []struct {
		name     string
		args     args
		want     want
		database database
	}{
		{
			name: "updates device wih specific ID",
			args: args{
				ctx: context.Background(),
				device: Device{
					ID:         1,
					Hostname:   "hostname1new",
					IPAddress:  "1.0.0.10",
					Login:      "user1new",
					Keyfile:    []byte{},
					Connected:  time.Date(2024, 5, 23, 0, 0, 0, 0, time.UTC),
					LastStatus: -1,
				},
			},
			want: want{
				devices: []Device{
					{
						ID:         1,
						Hostname:   "hostname1new",
						IPAddress:  "1.0.0.10",
						Login:      "user1new",
						Keyfile:    []byte{},
						Connected:  time.Date(2024, 5, 23, 0, 0, 0, 0, time.UTC),
						LastStatus: -1,
					},
					{
						ID:         2,
						Hostname:   "hostname2",
						IPAddress:  "10.0.0.2",
						Login:      "user2",
						Keyfile:    []byte{},
						Connected:  time.Date(2024, 5, 22, 0, 0, 0, 0, time.UTC),
						LastStatus: -1,
					},
				},
			},
			database: database{
				prepare: exec(`INSERT INTO devices(id, hostname, ip, login, connected)
VALUES (1,'hostname1','10.0.0.1','user1','2024-05-22 00:00:00'),
       (2,'hostname2','10.0.0.2','user2','2024-05-22 00:00:00');`),
				cleanup: cleanup("devices"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := connect()
			if err != nil {
				t.Fatalf("unable to connect to database: %v", err)
			}

			if tc.database.prepare != nil {
				tc.database.prepare(t, conn)
			}
			if tc.database.cleanup != nil {
				t.Cleanup(func() { tc.database.cleanup(t, conn) })
			}

			db := New(conn)

			err = db.UpdateDevice(tc.args.ctx, tc.args.device)

			errComp := gocmp.Comparer(func(x, y error) bool {
				return x.Error() == y.Error()
			})

			if diff := gocmp.Diff(err, tc.want.err, errComp); diff != "" {
				t.Errorf("error mismatch (-got +want):\n%s", diff)
			}

			got, _ := db.Devices(tc.args.ctx)

			deviceComp := gocmp.Comparer(func(x, y Device) bool {
				return x.ID == y.ID &&
					x.Hostname == y.Hostname &&
					x.IPAddress == y.IPAddress &&
					x.Login == y.Login &&
					x.Password == y.Password &&
					string(x.Keyfile) == string(y.Keyfile) &&
					x.Connected == y.Connected &&
					x.LastStatus == y.LastStatus
			})

			if diff := gocmp.Diff(got, tc.want.devices, deviceComp); diff != "" {
				t.Errorf("device read mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestDB_UpdateDeviceStatus(t *testing.T) {
	type args struct {
		ctx    context.Context
		device Device
	}
	type want struct {
		devices []Device
		err     error
	}
	type database struct {
		prepare func(*testing.T, *sql.DB)
		cleanup func(*testing.T, *sql.DB)
	}
	tests := []struct {
		name     string
		args     args
		want     want
		database database
	}{
		{
			name: "updates device wih specific ID",
			args: args{
				ctx: context.Background(),
				device: Device{
					ID:         1,
					Connected:  time.Date(2024, 5, 23, 0, 0, 0, 0, time.UTC).Add(time.Hour),
					LastStatus: 0,
				},
			},
			want: want{
				devices: []Device{
					{
						ID:         1,
						Hostname:   "hostname1",
						IPAddress:  "10.0.0.1",
						Login:      "user1",
						Keyfile:    []byte{},
						Connected:  time.Date(2024, 5, 23, 0, 0, 0, 0, time.UTC).Add(time.Hour),
						LastStatus: 0,
					},
					{
						ID:         2,
						Hostname:   "hostname2",
						IPAddress:  "10.0.0.2",
						Login:      "user2",
						Keyfile:    []byte{},
						Connected:  time.Date(2024, 5, 22, 0, 0, 0, 0, time.UTC),
						LastStatus: -1,
					},
				},
			},
			database: database{
				prepare: exec(`INSERT INTO devices(id, hostname, ip, login, connected)
VALUES (1,'hostname1','10.0.0.1','user1','2024-05-22 00:00:00'),
       (2,'hostname2','10.0.0.2','user2','2024-05-22 00:00:00');`),
				cleanup: cleanup("devices"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := connect()
			if err != nil {
				t.Fatalf("unable to connect to database: %v", err)
			}

			if tc.database.prepare != nil {
				tc.database.prepare(t, conn)
			}
			if tc.database.cleanup != nil {
				t.Cleanup(func() { tc.database.cleanup(t, conn) })
			}

			db := New(conn)

			err = db.UpdateDeviceStatus(tc.args.ctx, tc.args.device)

			errComp := gocmp.Comparer(func(x, y error) bool {
				return x.Error() == y.Error()
			})

			if diff := gocmp.Diff(err, tc.want.err, errComp); diff != "" {
				t.Errorf("error mismatch (-got +want):\n%s", diff)
			}

			got, _ := db.Devices(tc.args.ctx)

			deviceComp := gocmp.Comparer(func(x, y Device) bool {
				return x.ID == y.ID &&
					x.Hostname == y.Hostname &&
					x.IPAddress == y.IPAddress &&
					x.Login == y.Login &&
					x.Password == y.Password &&
					string(x.Keyfile) == string(y.Keyfile) &&
					x.Connected == y.Connected &&
					x.LastStatus == y.LastStatus
			})

			if diff := gocmp.Diff(got, tc.want.devices, deviceComp); diff != "" {
				t.Errorf("device read mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestDB_DeleteDevice(t *testing.T) {
	type args struct {
		ctx context.Context
		id  uint
	}
	type want struct {
		count *int
		err   error
	}
	type database struct {
		prepare func(*testing.T, *sql.DB)
		cleanup func(*testing.T, *sql.DB)
	}
	tests := []struct {
		name     string
		args     args
		want     want
		database database
	}{
		{
			name: "creates device",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			want: want{
				count: ptr(1),
			},
			database: database{
				prepare: exec(`INSERT INTO devices(id, hostname, ip, login, connected)
VALUES (1,'hostname1','10.0.0.1','user1','2024-05-22 00:00:00'),
       (2,'hostname2','10.0.0.2','user2','2024-05-22 00:00:00');`),
				cleanup: cleanup("devices"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := connect()
			if err != nil {
				t.Fatalf("unable to connect to database: %v", err)
			}

			if tc.database.prepare != nil {
				tc.database.prepare(t, conn)
			}
			if tc.database.cleanup != nil {
				t.Cleanup(func() { tc.database.cleanup(t, conn) })
			}

			db := New(conn)
			err = db.DeleteDevice(tc.args.ctx, tc.args.id)

			errComp := gocmp.Comparer(func(x, y error) bool {
				return x.Error() == y.Error()
			})

			if diff := gocmp.Diff(err, tc.want.err, errComp); diff != "" {
				t.Errorf("error mismatch (-got +want):\n%s", diff)
			}

			if tc.want.count != nil {
				got := count("devices")(t, conn)
				if diff := gocmp.Diff(got, *tc.want.count); diff != "" {
					t.Errorf("devices count mismatch (-got +want):\n%s", diff)
				}
			}
		})
	}
}
