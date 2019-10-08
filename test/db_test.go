package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/f6/webserver/webhandler"
	"github.com/fabarj4/example/simpleapi/model"
	_ "github.com/lib/pq"
)

var (
	user      = "postgres"
	password  = "postgres"
	dbtest    = "test_simple_api"
	defaultdb = "postgres"
)

// dbh provide db to handler.
type dbh struct {
	db *sql.DB
}

func (d dbh) Handle(w http.ResponseWriter, r *http.Request) webhandler.Response {
	res := webhandler.Response{}
	res.Ctx = webhandler.NewContextWithDB(r.Context(), d.db)
	return res
}

//table struct digunakan untuk membuat table
type table struct {
	Name       string
	Fields     []string
	PrimaryKey string
}

//connectDB digunakan untuk membuat koneksi dengan database
func connectDB(name, user, password string) (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable", user, name, password))
	return db, err
}

//createDB digunakan untuk membuat
func createDB(db *sql.DB, name string) error {
	query := fmt.Sprintf("CREATE DATABASE %s", name)
	_, err := db.Exec(query)
	return err
}

//createTable digunakan untuk membuat table pada database yang dipiih
func createTable(db *sql.DB, table table) error {
	var query string
	if table.PrimaryKey != "" {
		query = fmt.Sprintf("CREATE TABLE %s (%s, PRIMARY KEY %s)", table.Name, strings.Join(table.Fields, ","), table.PrimaryKey)
	} else {
		query = fmt.Sprintf("CREATE TABLE %s (%s)", table.Name, strings.Join(table.Fields, ","))
	}
	_, err := db.Exec(query)
	return err
}

//dropDB digunakan untuk menghapus DATABASE
func dropDB(db *sql.DB, name string) error {
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", name)
	_, err := db.Exec(query)
	return err
}

//PrepareTest digunakan untuk testing pembuatan, penghapusan datbaase dan table
func PrepareTest(t *testing.T) *sql.DB {
	t.Helper()
	//prepare mengahpus dan membuat database
	db, err := connectDB(defaultdb, user, password)
	if err != nil {
		t.Fatal(err)
	}
	if err := dropDB(db, dbtest); err != nil {
		t.Fatal(err)
	}
	if err = createDB(db, dbtest); err != nil {
		t.Fatal(err)
	}
	db.Close()
	//membuat table
	db, err = connectDB(dbtest, user, password)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(model.TBMahasiswa); err != nil {
		t.Fatal(err)
	}
	return db
}
