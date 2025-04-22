package graphql

import (
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

// Handler returns an HTTP handler for GraphQL requests
func Handler(resolver *Resolver) http.Handler {
	// Read the schema file
	schemaFile := filepath.Join("api", "graphql", "schema.graphql")
	schema, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		log.Fatalf("Failed to read schema: %v", err)
	}

	// Parse the schema
	parsedSchema := graphql.MustParseSchema(string(schema), resolver)

	// Create an HTTP handler
	return &relay.Handler{Schema: parsedSchema}
}
