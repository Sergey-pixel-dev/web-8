package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

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

func (h *Handlers) handler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	time, err := h.dbProvider.GetTimeLastVisit(name)
	if err == sql.ErrNoRows {
		err2 := h.dbProvider.SetTimeLastVisit(name)
		if err2 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("ошибка хахаххаха"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, " + name))
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ошибка хахаххаха"))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, " + name + " . your last visit was in " + time))
		err = h.dbProvider.UpdateTimeLastVisit(name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("ошибка хахаххаха"))
		}
	}
}

func (dp *DatabaseProvider) GetTimeLastVisit(name string) (string, error) {
	var msg string
	row := dp.db.QueryRow("SELECT time FROM labs where name=($1)", name) //чтоб наверняка последнее взяли
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}
	return msg, nil
}
func (dp *DatabaseProvider) UpdateTimeLastVisit(name string) error {
	_, err := dp.db.Exec("update labs set time = ($1) where name = ($2)", time.Now().Format("2006-01-02 15:04:05"), name)
	return err
}
func (dp *DatabaseProvider) SetTimeLastVisit(name string) error {
	_, err := dp.db.Exec("insert into labs (count, name, time) values (0, ($1), ($2))",
		name, time.Now().Format("2006-01-02 15:04:05"))
	return err
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

	http.HandleFunc("/api/user", h.handler)
	err = http.ListenAndServe(*address, nil)
	if err != nil {
		log.Fatal(err)
	}
}
