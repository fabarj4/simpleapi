package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/f6/webserver/webhandler"
	"github.com/fabarj4/example/simpleapi/handler"
	"github.com/fabarj4/example/simpleapi/model"
	"github.com/lib/pq"
)

var (
	db        *sql.DB
	dbgolang  = "simple_api"
	defaultdb = "postgres"
	dbuser    = "postgres"
	dbpass    = "postgres"

	notsecure = false
	port      = 8088
)

func init() {
	flag.BoolVar(&notsecure, "notsecure", false, "by default web server run on https, if notsecure true run in http")
	flag.IntVar(&port, "port", 8088, "port used in web sever")
	flag.StringVar(&dbuser, "dbuser", "postgres", "User for db postgres")
	flag.StringVar(&dbpass, "dbpass", "postgres", "Password for db postgres")
}

func main() {
	flag.Parse()
	var err error

	db, err = connectDB(dbgolang, dbuser, dbpass)
	if err != nil {
		if !isErrDBNotExist(err) {
			log.Fatalf("Gagal Konek database %s", err)
		}
		db, err = prepareDB()
		if err != nil {
			log.Fatal(err)
		}
	}

	webhandler.RegisterDB(db)
	webhandler.DebugOn()
	apiUrl := "/api/v1/testapi/"
	http.Handle(apiUrl, handler.WebHandler(apiUrl))
	fmt.Printf("Port used :%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func connectDB(name, user, password string) (*sql.DB, error) {
	// db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable", user, name, password))
	// kalo dipake menggunakan url
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", user, password, name))
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	return db, err
}

func createDB(db *sql.DB, name string) error {
	query := "CREATE DATABASE " + name
	_, err := db.Exec(query)
	return err
}

func isErrDBNotExist(err error) bool {
	if et, ok := err.(*pq.Error); ok {
		return et.Code == pq.ErrorCode("3D000")
	}
	return false
}

func prepareDB() (*sql.DB, error) {
	db, err := connectDB(defaultdb, dbuser, dbpass)
	if err != nil {
		return nil, err
	}

	if err = createDB(db, dbgolang); err != nil {
		return nil, err
	}
	db.Close()

	if db, err = connectDB(dbgolang, dbuser, dbpass); err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(model.TBMahasiswa); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// sample data
	data := []*model.Mahasiswa{
		&model.Mahasiswa{NPM: "53413109", Nama: "Faisal Akbar"},
		&model.Mahasiswa{NPM: "53413110", Nama: "Linda Asri Lelyandari"},
		&model.Mahasiswa{NPM: "53413111", Nama: "M Putera Yarman"},
	}
	for _, item := range data {
		tx, err := db.Begin()
		if err != nil {
			return nil, err
		}
		if err := item.Insert(tx); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}

	return db, nil
}
