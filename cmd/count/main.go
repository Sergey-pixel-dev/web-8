package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "bmstu"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

type envelope map[string]string

func (h *Handlers) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		encoder := json.NewDecoder(r.Body)
		i2 := envelope{"count": "0"}
		err := encoder.Decode(&i2)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("ошибка json"))
			return
		}
		i, err := strconv.Atoi(i2["count"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("это не число"))
			return
		}
		count, err := h.dbProvider.GetCount()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("ошибка в субд"))
			return
		}
		h.dbProvider.UpdateCount(count + i)
	}
	if r.Method == "GET" {
		count, err := h.dbProvider.GetCount()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("ошибка в субд"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))
	} else {

	}
}

func (dp *DatabaseProvider) GetCount() (int, error) {
	var msg string
	row := dp.db.QueryRow("SELECT count FROM labs order by id desc LIMIT 1") //чтоб наверняка последнее взяли
	err := row.Scan(&msg)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(msg)
}
func (dp *DatabaseProvider) UpdateCount(count int) error {
	_, err := dp.db.Exec("update labs set count = ($1)", count)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	address := flag.String("address", "127.0.0.1:8081", "адрес для запуска сервера")
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dp := DatabaseProvider{db: db}
	h := Handlers{dbProvider: dp}

	http.HandleFunc("/", h.handler)
	http.ListenAndServe(*address, nil)
}
