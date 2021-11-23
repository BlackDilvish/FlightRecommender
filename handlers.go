package main

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

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
