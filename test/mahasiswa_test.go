package main

import (
	"testing"

	"github.com/fabarj4/example/simpleapi/model"
)

func TestMahasiswa(t *testing.T) {
	db := PrepareTest(t)
	defer db.Close()
	data := []*model.Mahasiswa{
		&model.Mahasiswa{NPM: "53413109", Nama: "Faisal Akbar"},
		&model.Mahasiswa{NPM: "53413110", Nama: "Linda Asri Lelyandari"},
		&model.Mahasiswa{NPM: "53413111", Nama: "M Putera Yarman"},
	}
	t.Run("Test Insert Mahasiswa", func(t *testing.T) {
		for _, item := range data {
			fields, _ := item.Fields()
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			if err := item.Insert(tx); err != nil {
				t.Fatalf("Insert data : %v error : %v", item, err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			got := &model.Mahasiswa{NPM: item.NPM}
			if err := got.Get(db); err != nil {
				t.Fatal(err)
			}
			CompareMahasiswa(t, got, item, fields)
		}
	})

	t.Run("test Update Mahasiswa", func(t *testing.T) {
		changes := []map[string]interface{}{
			{"nama": "Verudi"},
		}
		dataUpdate := data[0]
		for _, change := range changes {
			_, err := dataUpdate.Update(db, change)
			if err != nil {
				t.Fatalf("Update error : %v", err)
			}
			got := &model.Mahasiswa{NPM: data[0].NPM}
			if err := got.Get(db); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("Test Delete Mahasiswa", func(t *testing.T) {
		for _, item := range data {
			if err := item.Delete(db); err != nil {
				t.Errorf("Delete npm : %v error : %v", item.NPM, err)
			}
		}
	})
}

func CompareMahasiswa(t *testing.T, got, want *model.Mahasiswa, fields []string) {
	if len(fields) == 0 {
		fields, _ = got.Fields()
	}
	for _, field := range fields {
		if field == "npm" && got.NPM != want.NPM {
			t.Errorf("Got NPM : %v want : %v", got.NPM, want.NPM)
		}
		if field == "nama" && got.Nama != want.Nama {
			t.Errorf("Got Nama : %v want : %v", got.Nama, want.Nama)
		}
	}
}
