package main

import (
	"net/http"
	"text/template"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var templates = template.Must(template.ParseFiles("templates/airport_form.html",
	"templates/home.html",
	"templates/navbar.tmpl.html",
	"templates/view.html",
	"templates/path_form.html",
	"templates/connection_form.html",
	"templates/country_form.html"))

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

func countryFormHandler(w http.ResponseWriter, r *http.Request, driver neo4j.Driver) {
	countries := readData(r, driver, getCountries)
	err := templates.ExecuteTemplate(w, "country_form.html", countries)
	if err != nil {
		return
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "home.html", "test")
	if err != nil {
		return
	}
}
