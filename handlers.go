package main

import (
	"strings"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func getAirport(tx neo4j.Transaction, vars map[string]string) (neo4j.Result, error) {
	key := vars["name"]

	cypher := `MATCH (a:Airport)
               WHERE a.name = $airport_name
               RETURN a.name, a.country`

	return tx.Run(cypher, map[string]interface{}{
		"airport_name": key,
	})
}

func getAirports(tx neo4j.Transaction, vars map[string]string) (neo4j.Result, error) {
	cypher := `MATCH (a:Airport)
               RETURN a.name, a.country`

	return tx.Run(cypher, map[string]interface{}{})
}

func getAirportsByCountry(tx neo4j.Transaction, vars map[string]string) (neo4j.Result, error) {
	key := strings.ReplaceAll(vars["country"], "%2B", "+")

	cypher := `MATCH (a:Airport)
               WHERE a.country = $airport_country
               RETURN a.name, a.country`

	return tx.Run(cypher, map[string]interface{}{
		"airport_country": key,
	})
}

func getCountries(tx neo4j.Transaction, vars map[string]string) (neo4j.Result, error) {
	cypher := `MATCH (a:Airport)
               WITH DISTINCT a.country AS country
               RETURN "null", country`

	return tx.Run(cypher, map[string]interface{}{})
}

func getConnectedAirportsOut(tx neo4j.Transaction, vars map[string]string) (neo4j.Result, error) {
	key := vars["name"]

	cypher := `MATCH (a:Airport)-[r:HAS_CONNECTION]->(b:Airport)
               WHERE a.name = $airport_name
               RETURN b.name, b.country`

	return tx.Run(cypher, map[string]interface{}{
		"airport_name": key,
	})
}

func getConnectedAirportsIn(tx neo4j.Transaction, vars map[string]string) (neo4j.Result, error) {
	key := vars["name"]

	cypher := `MATCH (a:Airport)-[r:HAS_CONNECTION]->(b:Airport)
               WHERE b.name = $airport_name
               RETURN a.name, a.country`

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
               Where a.name = $dept and b.name = $dest and a.name <> b.name and length(p) > 0
               UNWIND [n in nodes(p) | n] AS n
               RETURN n.name, n.country`

	return tx.Run(cypher, map[string]interface{}{
		"dept": dept,
		"dest": dest,
	})
}

func createAirport(tx neo4j.Transaction, reqBody []byte) (neo4j.Result, error) {
	var airport Airport
	args := parseRequestBody(string(reqBody))
	airport.Name = args["Name"]
	airport.Country = args["Country"]

	cypher := `CREATE (a:Airport { name: $name, country: $country })
	           RETURN a.name, a.country`

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
