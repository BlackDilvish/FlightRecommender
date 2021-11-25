package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func readHandler(w http.ResponseWriter, r *http.Request, driver neo4j.Driver, fn func(neo4j.Transaction, map[string]string) (neo4j.Result, error)) {
	records := readData(r, driver, fn)

	err := templates.ExecuteTemplate(w, "view.html", records)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request, driver neo4j.Driver, fn func(neo4j.Transaction, []byte) (neo4j.Result, error)) {
	err := saveData(r, driver, fn)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/airports", http.StatusFound)
}

func readData(r *http.Request, driver neo4j.Driver, fn func(neo4j.Transaction, map[string]string) (neo4j.Result, error)) []Airport {
	var records []Airport

	vars := mux.Vars(r)
	if len(vars) == 0 {
		bytes, _ := ioutil.ReadAll(r.Body)
		vars = parseRequestBody(string(bytes))
	}

	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()
	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {

		result, err := fn(tx, vars)

		if err != nil {
			return nil, err
		}
		for result.Next() {
			records = append(records, Airport{
				Name: result.Record().Values[0].(string),
			})
		}
		return nil, result.Err()
	})
	if err != nil {
		panic(err)
	}

	return records
}

func saveData(r *http.Request, driver neo4j.Driver, fn func(neo4j.Transaction, []byte) (neo4j.Result, error)) error {
	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()
	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		result, err := fn(tx, reqBody)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		return nil, result.Err()
	})
	if err != nil {
		return err
	}

	return nil
}

func parseRequestBody(body string) map[string]string {
	m := make(map[string]string)
	if body == "" {
		return m
	}

	for _, obj := range strings.Split(body, "&") {
		vars := strings.Split(obj, "=")
		m[vars[0]] = vars[1]
	}

	return m
}
