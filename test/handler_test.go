package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/f6/dba"
	"github.com/f6/webserver/webhandler"
	"github.com/fabarj4/example/simpleapi/handler"
	"github.com/fabarj4/example/simpleapi/model"
)

func TestHandler(t *testing.T) {
	db := PrepareTest(t)
	defer db.Close()
	handle := webhandler.New(dbh{db: db}, handler.TestAPIHandler{pattern: "/api/v1/testapi/"})
	ts := httptest.NewServer(handle)
	defer ts.Close()

	data := []*model.Mahasiswa{
		&model.Mahasiswa{NPM: "53413109", Nama: "Faisal Akbar"},
		&model.Mahasiswa{NPM: "53413110", Nama: "Linda Asri Lelyandari"},
		&model.Mahasiswa{NPM: "53413111", Nama: "M Putera Yarman"},
	}

	t.Run("Test Handler Insert", func(t *testing.T) {
		for _, item := range data {
			body, err := json.MarshalIndent(item, "", " ")
			if err != nil {
				t.Fatal(err)
			}
			client := &http.Client{}
			urls := ts.URL + "/api/v1/testapi/mahasiswa"
			req, err := http.NewRequest(http.MethodPost, urls, bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			res, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode != http.StatusOK {
				t.Fatalf("Got status : %v Want : %v", res.StatusCode, http.StatusOK)
			}
			defer res.Body.Close()
			dataBody, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Fatalf("Error read response body : %v", err)
			}

			type trd struct {
				Err    string          `jsin:"err"`
				Data   model.Mahasiswa `json:"data"`
				Cursor dba.Cursor      `json:"cursor"`
			}

			var rd trd

			if err = json.Unmarshal(dataBody, &rd); err != nil {
				t.Fatalf("err : %v data : %v", err, dataBody)
			}
			item.NPM = rd.Data.NPM
			got := &rd.Data
			fields, _ := item.Fields()
			CompareMahasiswa(t, got, item, fields)
		}
	})

	t.Run("Test Handler Gets", func(t *testing.T) {
		client := &http.Client{}
		urls := ts.URL + "/api/v1/testapi/mahasiswa"
		req, err := http.NewRequest(http.MethodGet, urls, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ""))
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got status : %v Want : %v", res.StatusCode, http.StatusOK)
		}
		dataBody, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			t.Fatalf("Error read response body : %v", err)
		}

		type trd struct {
			Err  string            `jsin:"err"`
			Data []model.Mahasiswa `json:"data"`
		}

		var rd trd

		if err = json.Unmarshal(dataBody, &rd); err != nil {
			t.Fatalf("err : %v data : %v", err, dataBody)
		}

		got := rd.Data
		if len(got) != len(data) {
			t.Errorf("Got data : %d Want : %d\n", len(got), len(data))
		}
	})

	t.Run("Test Handler Get", func(t *testing.T) {
		client := &http.Client{}
		urls := ts.URL + "/api/v1/testapi/mahasiswa/" + data[0].NPM
		req, err := http.NewRequest(http.MethodGet, urls, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ""))
		res, err := client.Do(req)

		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got status : %v Want : %v", res.StatusCode, http.StatusOK)
		}
		dataBody, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatalf("Error read response body : %v", err)
		}

		type trd struct {
			Err    string          `jsin:"err"`
			Data   model.Mahasiswa `json:"data"`
			Cursor dba.Cursor      `json:"cursor"`
		}

		var rd trd

		if err = json.Unmarshal(dataBody, &rd); err != nil {
			t.Fatalf("err : %v data : %v", err, dataBody)
		}

		got := &rd.Data
		fields, _ := data[0].Fields()
		CompareMahasiswa(t, got, data[0], fields)
	})

	t.Run("Test Handler Filter", func(t *testing.T) {
		urls := ts.URL + "/api/v1/testapi/mahasiswa/"
		var filter = []struct {
			fil  string
			want []*model.Mahasiswa
		}{
			{
				fil:  "npm,=,53413109",
				want: data[0:1],
			},
		}
		for _, item := range filter {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?filters=%s", urls, url.QueryEscape(item.fil)), nil)
			if err != nil {
				t.Fatalf("error get filter : %v", err)
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ""))
			client := http.Client{}
			res, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()

			dataBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Error read response body : %v", err)
			}

			type trd struct {
				Err  string             `jsin:"err"`
				Data []*model.Mahasiswa `json:"data"`
			}

			var rd trd

			if err = json.Unmarshal(dataBody, &rd); err != nil {
				t.Fatalf("err : %v data : %v", err, dataBody)
			}

			got := rd.Data
			if len(got) != len(item.want) {
				t.Errorf("Got data : %d Want : %d\n", len(got), len(data))
			}
		}
	})

	t.Run("Test Handler Update", func(t *testing.T) {
		dataUpdate := map[string]interface{}{
			"npm": "53413109", "nama": "falbar",
		}
		jsonUpdate, err := json.MarshalIndent(dataUpdate, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		urls := fmt.Sprintf("%s/api/v1/testapi/mahasiswa/%s", ts.URL, data[0].NPM)
		req, err := http.NewRequest(http.MethodPut, urls, bytes.NewBuffer(jsonUpdate))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ""))
		client := http.Client{}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got Status : %v Want : %v", res.StatusCode, http.StatusOK)
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		type trd struct {
			Err  string           `json:"err"`
			Data *model.Mahasiswa `json:"data"`
		}
		var rd trd
		if err := json.Unmarshal(body, &rd); err != nil {
			t.Fatalf("Err : %v Data : %s", err, body)
		}
		got := &model.Mahasiswa{NPM: data[0].NPM}
		err = dba.Get(db, got, "")
		if err != nil {
			t.Fatalf("Get Data : %v err : %v", data[0], err)
		}
		fmt.Println(rd)
	})

	t.Run("Test Handler Delete", func(t *testing.T) {
		urls := ts.URL + "/api/v1/testapi/mahasiswa/" + data[0].NPM
		// b, err := json.MarshalIndent(dataDelete, "", " ")
		req, err := http.NewRequest(http.MethodDelete, urls, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ""))
		client := http.Client{}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got Status : %v Want : %v", res.StatusCode, http.StatusOK)
		}
		defer res.Body.Close()
		got := &model.Mahasiswa{NPM: data[0].NPM}
		err = dba.Get(db, got, "")
		if err == nil {
			t.Fatalf("Get Data : %v  Err : Data Tidak Terhapus", got)
		}
	})
}
