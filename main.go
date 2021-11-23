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
	"templates/header.html",
	"templates/path_form.html",
	"templates/connection_form.html"))

func readHandler(w http.ResponseWriter, r *http.Request, driver neo4j.Driver, fn func(neo4j.Transaction, map[string]string) (neo4j.Result, error)) {
	records := readData(r, driver, fn)

	err := templates.ExecuteTemplate(w, "view.html", records)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	dept := vars["departure"]
	dest := vars["destination"]

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

func createConnection(tx neo4j.Transaction, reqBody []byte) (neo4j.Result, error) {
	var connection Connection
	args := parseRequestBody(string(reqBody))
	connection.Departure = args["departure"]
	connection.Destination = args["destination"]

	cypher := `MATCH (a:Airport { name: $dept })
	           MATCH (b:Airport { name: $dest })
	           CREATE (a)-[rel:HAS_CONNECTION]->(b)`

	return tx.Run(cypher, map[string]interface{}{
		"dept": connection.Departure,
		"dest": connection.Destination,
	})
}

func airportFormHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "airport_form.html", "test")
	if err != nil {
		return
	}
}

func pathFormHandler(w http.ResponseWriter, r *http.Request, driver neo4j.Driver) {
	airports := readData(r, driver, getAirports)
	err := templates.ExecuteTemplate(w, "path_form.html", airports)
	if err != nil {
		return
	}
}

func connectionFormHandler(w http.ResponseWriter, r *http.Request, driver neo4j.Driver) {
	airports := readData(r, driver, getAirports)
	err := templates.ExecuteTemplate(w, "connection_form.html", airports)
	if err != nil {
		return
	}
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

func handleRequests(driver neo4j.Driver) {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/airports", func(w http.ResponseWriter, r *http.Request) { readHandler(w, r, driver, getAirports) })
	myRouter.HandleFunc("/airport/{name}", func(w http.ResponseWriter, r *http.Request) { readHandler(w, r, driver, getAirport) })
	myRouter.HandleFunc("/connections/{name}", func(w http.ResponseWriter, r *http.Request) { readHandler(w, r, driver, getConnectedAirports) })

	myRouter.HandleFunc("/airport", airportFormHandler).Methods("GET")
	myRouter.HandleFunc("/airport", func(w http.ResponseWriter, r *http.Request) { saveData(w, r, driver, createAirport) }).Methods("POST")

	myRouter.HandleFunc("/connection", func(w http.ResponseWriter, r *http.Request) { connectionFormHandler(w, r, driver) }).Methods("GET")
	myRouter.HandleFunc("/connection", func(w http.ResponseWriter, r *http.Request) { saveData(w, r, driver, createConnection) }).Methods("POST")

	myRouter.HandleFunc("/path", func(w http.ResponseWriter, r *http.Request) { pathFormHandler(w, r, driver) }).Methods("GET")
	myRouter.HandleFunc("/path", func(w http.ResponseWriter, r *http.Request) { readHandler(w, r, driver, getPath) }).Methods("POST")

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

type Connection struct {
	Departure   string
	Destination string
}
