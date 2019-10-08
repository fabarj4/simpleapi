package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/f6/dba"
	"github.com/f6/webserver/webhandler"
	"github.com/fabarj4/example/simpleapi/model"
)

type TestAPIHandler struct {
	Pattern string
}

func UrlPath(u *url.URL, pattern string) []string {
	urlpath := u.RawPath
	if urlpath == "" {
		urlpath = u.Path
	}
	pathpattern := strings.TrimPrefix(urlpath, pattern)
	path := strings.Split(pathpattern, "/")
	return path
}

//QueryFields
func QueryFields(query url.Values) ([]string, error) {
	qry := query.Get("fields")
	var flds []string
	if qry != "" {
		flds = strings.Split(qry, ",")

	}
	return flds, nil
}

//QueryCursor
func QueryCursor(query url.Values) (set bool, cursor dba.Cursor, err error) {
	qry := query.Get("cursor")
	if qry != "" {
		cursor, err = dba.Decode(qry)
		set = true
	}
	return set, cursor, err
}

//QueryLimit
func QueryLimit(query url.Values) (int, error) {
	qry := query.Get("limit")
	var lmt int
	var err error
	if qry != "" {
		lmt, err = strconv.Atoi(qry)
		if err != nil {
			return 0, err
		}
	}
	return lmt, nil
}

//QuerySort
func QuerySort(query url.Values) ([]string, bool, error) {
	qry := query.Get("sort")
	var srt []string
	descending := true
	if qry != "" {
		fieldsrt := strings.Split(qry, " ")
		srt = strings.Split(fieldsrt[0], ",")
		if len(srt) < 1 {
			return nil, false, fmt.Errorf("tidak memenuhi persyaratan")
		}
		if fieldsrt[1] != "DESC" {
			descending = false
		}
	}
	return srt, descending, nil
}

//QueryFilter
func QueryFilter(query url.Values) ([]dba.Filter, error) {
	qry := query.Get("filters")

	var fil []dba.Filter
	if qry != "" {

		fieldspr := strings.Split(qry, ";")

		for _, tc := range fieldspr {
			parameter := strings.Split(tc, ",")
			if len(parameter) != 3 {
				return nil, fmt.Errorf("tidak memenuhi persyaratan")
			}
			valueparameter, err := url.PathUnescape(parameter[2])
			// fmt.Printf("id : %s", id)
			if err != nil {
				return nil, fmt.Errorf("tidak memenuhi persyaratan")
			}
			b := dba.Filter{
				Field: parameter[0],
				Op:    parameter[1],
				Value: valueparameter,
			}
			fil = append(fil, b)
		}
	}
	return fil, nil
}
func newTx(db *sql.DB) *sql.Tx {
	tx, err := db.Begin()
	if err != nil {
		return nil
	}
	return tx
}

func WebHandler(pattern string) http.Handler {
	wh := webhandler.New(TestAPIHandler{Pattern: pattern})
	return wh
}

func (h TestAPIHandler) Handle(w http.ResponseWriter, r *http.Request) webhandler.Response {
	var res webhandler.Response

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/testapi/")
	splitPath := strings.Split(path, "/")
	urls := splitPath[0]
	switch {
	case urls == "mahasiswa":
		switch {
		case r.Method == http.MethodPut:
			res = h.handlePut(w, r)
		case r.Method == http.MethodPost:
			res = h.handlePost(w, r)
		case r.Method == http.MethodDelete:
			res = h.handleDelete(w, r)
		case r.Method == http.MethodGet:
			res = h.handleGet(w, r)
		}
	default:
		res.ErrMessage = "Not Found"
		return res
	}
	return res
}

func (h TestAPIHandler) handlePost(w http.ResponseWriter, r *http.Request) webhandler.Response {
	res := webhandler.Response{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res.Error(err, http.StatusBadRequest)
		return res
	}
	mahasiswa := &model.Mahasiswa{}
	if err := json.Unmarshal(body, mahasiswa); err != nil {
		res.Error(err, http.StatusBadRequest)
		return res
	}
	db, err := webhandler.DBFromContext(r.Context())
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	tx, err := db.Begin()
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	if err = mahasiswa.Insert(tx); err != nil {
		res.Error(err, http.StatusOK)
		tx.Rollback()
		return res
	}
	if err = tx.Commit(); err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	res.Data = mahasiswa
	return res
}

func (h TestAPIHandler) handleGet(w http.ResponseWriter, r *http.Request) webhandler.Response {
	res := webhandler.Response{}
	paths := UrlPath(r.URL, h.Pattern)

	db, err := webhandler.DBFromContext(r.Context())
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	if len(paths) == 2 {
		id, err := url.PathUnescape(paths[1])
		if err != nil {
			res.Error(err, http.StatusInternalServerError)
			return res
		}
		if id != "" {
			// idInt, _ := strconv.Atoi(id)
			m := &model.Mahasiswa{NPM: id}
			err := m.Get(db)
			if err != nil {
				res.Error(err, http.StatusOK)
				return res
			}
			res.Data = m
			return res
		}
	}
	var values = r.URL.Query()
	set, cursor, err := QueryCursor(values)
	if err != nil {
		res.Error(err, http.StatusBadRequest)
	}
	if !set {
		fields, err := QueryFields(values)
		if err != nil {
			res.Error(err, http.StatusBadRequest)
			return res
		}
		limit, err := QueryLimit(values)
		if err != nil {
			res.Error(err, http.StatusBadRequest)
			return res
		}
		if limit == 0 {
			limit = 1000
		}
		sort, asc, err := QuerySort(values)
		if err != nil {
			res.Error(err, http.StatusBadRequest)
			return res
		}
		filter, err := QueryFilter(values)
		if err != nil {
			res.Error(err, http.StatusBadRequest)
			return res
		}
		cursor = dba.Cursor{
			Fields:     fields,
			Filters:    filter,
			OrderBy:    sort,
			Descending: asc,
			Limit:      limit,
		}
	}

	var data []dba.Table
	m := &model.Mahasiswa{}
	data, cursor, err = dba.Fetch(db, m, "", cursor)
	if err != nil {
		res.Error(err, http.StatusOK)
		return res
	}
	res.Data = data
	res.Cursor = cursor
	return res
}

func (h TestAPIHandler) handleDelete(w http.ResponseWriter, r *http.Request) webhandler.Response {
	res := webhandler.Response{}
	paths := UrlPath(r.URL, h.Pattern)
	last, err := url.PathUnescape(paths[1])
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	if last == "" || last == "mahasiswa" {
		res.Error(errors.New("npm tidak boleh kosong"), http.StatusOK)
		return res
	}
	db, err := webhandler.DBFromContext(r.Context())
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	tx, err := db.Begin()
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	mahasiswa := model.Mahasiswa{NPM: last}
	if err = mahasiswa.Get(db); err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	if err = mahasiswa.Delete(tx); err != nil {
		res.Error(err, http.StatusOK)
		tx.Rollback()
		return res
	}
	if err = tx.Commit(); err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	return res
}

func (h TestAPIHandler) handlePut(w http.ResponseWriter, r *http.Request) webhandler.Response {
	res := webhandler.Response{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res.Error(err, http.StatusBadRequest)
		return res
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		res.Error(err, http.StatusBadRequest)
		return res
	}

	paths := UrlPath(r.URL, h.Pattern)
	last, err := url.PathUnescape(paths[1])
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	if last == "" || last == "mahasiswa" {
		res.Error(errors.New("npm tidak boleh kosong"), http.StatusOK)
		return res
	}

	db, err := webhandler.DBFromContext(r.Context())
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	tx, err := db.Begin()
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	// idTemp, _ := strconv.Atoi()
	// m := &PSPDetail{ID: fmt.Sprintf("%s", id)}
	mahasiswa := model.Mahasiswa{NPM: last}
	if err = mahasiswa.Get(db); err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	data, err = mahasiswa.Update(tx, data)
	if err != nil {
		tx.Rollback()
		res.Error(err, http.StatusOK)
		return res
	}
	if err = tx.Commit(); err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	res.Data = data
	return res
}
