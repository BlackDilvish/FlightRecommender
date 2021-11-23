package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var templates = template.Must(template.ParseFiles("templates/airport_form.html",
	"templates/view.html",
	"templates/header.html"))

func readData(w http.ResponseWriter, r *http.Request, driver neo4j.Driver, fn func(neo4j.Transaction, map[string]string) (neo4j.Result, error)) {
	var records []Airport

	vars := mux.Vars(r)

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

	err = templates.ExecuteTemplate(w, "view.html", records)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getAirport(tx neo4j.Transaction, vars map[string]string) (neo4j.Result, error) {
	key := vars["name"]

	cypher := `MATCH (a:Airport)
               WHERE a.name = $airport_name
               RETURN a.name AS name`

	return tx.Run(cypher, map[string]interface{}{
		"airport_name": key,
	})
}

func getAirports(tx neo4j.Transaction, vars map[string]string) (neo4j.Result, error) {

	cypher := `MATCH (a:Airport)
               RETURN a.name AS name`

	return tx.Run(cypher, map[string]interface{}{})
}

func getConnectedAirports(tx neo4j.Transaction, vars map[string]string) (neo4j.Result, error) {
	key := vars["name"]

	cypher := `MATCH (a:Airport)-[r:HAS_CONNECTION]->(b:Airport)
               WHERE a.name = $airport_name
               RETURN b.name`

	return tx.Run(cypher, map[string]interface{}{
		"airport_name": key,
	})
}

func getPath(tx neo4j.Transaction, vars map[string]string) (neo4j.Result, error) {
	dept := vars["dept"]
	dest := vars["dest"]

	cypher := `MATCH (a:Airport),
               (b:Airport),
               p = shortestPath((a)-[HAS_CONNECTION*]->(b))
               Where a.name = $dept and b.name = $dest and length(p) > 0
               UNWIND [n in nodes(p) | n.name] as name
               RETURN name`

	return tx.Run(cypher, map[string]interface{}{
		"dept": dept,
		"dest": dest,
	})
}

func saveData(w http.ResponseWriter, r *http.Request, driver neo4j.Driver, fn func(neo4j.Transaction, []byte) (neo4j.Result, error)) error {
	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()
	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		result, err := createAirport(tx, reqBody)

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

func createAirport(tx neo4j.Transaction, reqBody []byte) (neo4j.Result, error) {
	var airport Airport
	args := parseRequestBody(string(reqBody))
	airport.Name = args["Name"]
	airport.Country = args["Country"]

	cypher := `CREATE (a:Airport { name: $name, country: $country })
	           RETURN a.name`

	return tx.Run(cypher, map[string]interface{}{
		"name":    airport.Name,
		"country": airport.Country,
	})
}

func airportFormHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "airport_form.html", "test")
	if err != nil {
		return
	}
}

func parseRequestBody(body string) map[string]string {
	m := make(map[string]string)
	for _, obj := range strings.Split(body, "&") {
		vars := strings.Split(obj, "=")
		m[vars[0]] = vars[1]
	}

	return m
}

func handleRequests(driver neo4j.Driver) {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/airports", func(w http.ResponseWriter, r *http.Request) { readData(w, r, driver, getAirports) })
	myRouter.HandleFunc("/airport/{name}", func(w http.ResponseWriter, r *http.Request) { readData(w, r, driver, getAirport) })
	myRouter.HandleFunc("/connections/{name}", func(w http.ResponseWriter, r *http.Request) { readData(w, r, driver, getConnectedAirports) })
	myRouter.HandleFunc("/path/{dept}/{dest}", func(w http.ResponseWriter, r *http.Request) { readData(w, r, driver, getPath) })

	myRouter.HandleFunc("/create/airport", airportFormHandler)
	myRouter.HandleFunc("/airport", func(w http.ResponseWriter, r *http.Request) { saveData(w, r, driver, createAirport) }).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, myRouter))
}

func main() {

	uri := "neo4j+s://5ac579fa.databases.neo4j.io"
	pass := os.Getenv("FLIGHT_PASS")
	auth := neo4j.BasicAuth("neo4j", pass, "")

	driver, err := neo4j.NewDriver(uri, auth)
	if err != nil {
		panic(err)
	}
	defer driver.Close()

	handleRequests(driver)
}

type Airport struct {
	Name    string
	Country string
}
