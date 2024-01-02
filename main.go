package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Book struct {
	Title       string
	Description string
	Price       string
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "doni"
	dbname   = "postgres"
)

func func1(w http.ResponseWriter, dec *json.Decoder, b Book) {
	err := dec.Decode(&b)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			http.Error(w, msg, http.StatusBadRequest)

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		default:
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
}

func CreateBook(w http.ResponseWriter, r *http.Request) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, _ := sql.Open("postgres", psqlInfo)

	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var b Book
	

	func1(w, dec, b)

	err3 := dec.Decode(&struct{}{})
	if !errors.Is(err3, io.EOF) {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	sqlStatement := `
	INSERT INTO book (title, description, price)
	VALUES ($1, $2, $3)
	RETURNING *`

	_, err4 := db.Exec(sqlStatement, b.Title, b.Description, b.Price)
	if err4 != nil {
		log.Fatal(err4)
	}

	fmt.Fprintf(w, "Person: %+v", w)
}

func GetAll(w http.ResponseWriter, r *http.Request) {
	
	// func1(w, dec)
}

func main() {
	r := mux.NewRouter()
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	r.HandleFunc("/book/create", CreateBook)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":4000", r)

	err3 := db.Ping()
	if err3 != nil {
		panic(err3)
	}
	defer db.Close()
}
