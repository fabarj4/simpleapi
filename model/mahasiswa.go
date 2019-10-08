package model

import (
	"fmt"

	"github.com/f6/dba"
)

var TBMahasiswa = `
  CREATE TABLE mahasiswa
  (
    npm varchar(10),
    nama varchar(50)
  );
`

type Mahasiswa struct {
	NPM  string `json:"npm"`
	Nama string `json:"nama"`
}

func (m *Mahasiswa) Name() string {
	return "mahasiswa"
}

func (m *Mahasiswa) Fields() (fields []string, dst []interface{}) {
	fields = []string{"npm", "nama"}
	dst = []interface{}{&m.NPM, &m.Nama}
	return fields, dst
}

func (m *Mahasiswa) PrimaryKey() (fields []string, dst []interface{}) {
	fields = []string{"npm"}
	dst = []interface{}{&m.NPM}
	return fields, dst
}

func (m *Mahasiswa) HasAutoIncrementField() bool {
	return false
}

func (m *Mahasiswa) New() dba.Table {
	return &Mahasiswa{}
}

func (m *Mahasiswa) Insert(db dba.DBExecer) error {
	if m.NPM == "" {
		return fmt.Errorf("npm tidak boleh kosong")
	}
	return dba.Insert(db, m, "")
}

func (m *Mahasiswa) Update(db dba.DBExecer, change map[string]interface{}) (map[string]interface{}, error) {
	return change, dba.Update(db, m, "", change)
}

func (m *Mahasiswa) Delete(db dba.DBExecer) error {
	return dba.Delete(db, m, "")
}

func (m *Mahasiswa) Get(db dba.DBExecer) error {
	return dba.Get(db, m, "")
}

func Mahasiswas(db dba.DBExecer, c dba.Cursor) ([]*Mahasiswa, dba.Cursor, error) {
	mhs := &Mahasiswa{}
	res, cur, err := dba.Fetch(db, mhs, "", c)
	if err != nil {
		return nil, c, err
	}
	result := make([]*Mahasiswa, len(res))
	for i, v := range res {
		result[i] = v.(*Mahasiswa)
	}
	return result, cur, nil
}
