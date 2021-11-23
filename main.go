package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

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
