package graphql

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

//go:embed *.graphql
var content embed.FS

func readSchema() (string, error) {
	var buf bytes.Buffer

	fn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walking dir: %w", err)
		}

		// Only add files with the .graphql extension.
		if !strings.HasSuffix(path, ".graphql") {
			return nil
		}

		b, err := content.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading file %q: %w", path, err)
		}

		// Add a newline to separate each file.
		b = append(b, []byte("\n")...)

		if _, err := buf.Write(b); err != nil {
			return fmt.Errorf("writing %q bytes to buffer: %w", path, err)
		}

		return nil
	}

	// Recursively walk this directory and append all the file contents together.
	if err := fs.WalkDir(content, ".", fn); err != nil {
		return buf.String(), fmt.Errorf("walking content directory: %w", err)
	}

	return buf.String(), nil
}

// Handler returns an HTTP handler for GraphQL requests
func Handler(resolver *Resolver) http.Handler {
	// Read the schema file
	schema, err := readSchema()

	if err != nil {
		log.Fatalf("Failed to read schema: %v", err)
	}

	fmt.Println(string(schema))

	// Parse the schema
	parsedSchema := graphql.MustParseSchema(string(schema), resolver)

	// Create an HTTP handler
	return &relay.Handler{Schema: parsedSchema}
}
