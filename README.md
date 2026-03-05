# MiniAPI

WARNING (This project is highly developing, please don't use it in any production application.)

MiniAPI is a lightweight, type-safe, and interface-driven API framework for Go 1.25+. It leverages Go generics to provide a developer-friendly experience for defining routes, automatically binding requests, and marshaling responses, while seamlessly integrating with OpenAPI specification generation.

## Features

*   **Type-Safe Handlers:** Define your routes using concrete Go structs for requests and responses, eliminating boilerplate type assertions and parsing.
*   **Automatic Binding:** Automatically bind HTTP requests (path parameters, query strings, headers, and JSON body) to your request structs using struct tags.
*   **OpenAPI Integration:** Automatically generate OpenAPI v3 documentation based on your defined routes, groups, and structures.
*   **Framework Agnostic (Adapters):** Built around core interfaces, allowing you to plug MiniAPI into existing web frameworks. An adapter for `echo/v4` is included out-of-the-box.
*   **Routing & Groups:** Organize your API with logical groups, shared prefixes, and route-specific or group-wide middlewares.

## Installation

```bash
go get github.com/cirius-go/miniapi
```

## Quick Start

Here is a quick example of how to use MiniAPI with the Echo v4 framework.

```go
package main

import (
	"context"
	"net/http"

	"github.com/cirius-go/miniapi"
	"github.com/cirius-go/miniapi/adapter/echov4"
	"github.com/cirius-go/miniapi/group"
	"github.com/cirius-go/miniapi/route"
	"github.com/labstack/echo/v4"
)

// Define your request and response models
type HelloRequest struct {
	Name string `query:"name" json:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

// Implement your type-safe handler
func HelloHandler(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	name := req.Name
	if name == "" {
		name = "World"
	}
	return &HelloResponse{
		Message: "Hello, " + name + "!",
	}, nil
}

func main() {
	// 1. Setup the underlying framework (Echo v4)
	e := echo.New()
	adapter := echov4.NewAdapter(e)

	// 2. Create a MiniAPI group
	rootGroup := group.New("/api")

	// 3. Define a type-safe route
	helloRoute := route.New(route.Spec{
		Method:  http.MethodGet,
		Path:    "/hello",
		ID:      "helloWorld",
		Summary: "Says hello",
	}, HelloHandler)

	// 4. Add the route to the group
	rootGroup.AddRoutes(helloRoute)

	// 5. Register the group with the adapter
	// Typically you would iterate through the group's routes and register them
	for r := range rootGroup.Routes() {
		adapter.AddRoute(rootGroup.Path(), r)
	}

	// Start the server
	e.Logger.Fatal(e.Start(":8080"))
}
```

## Architecture

MiniAPI is built upon several core interfaces defined in `contract.go`:

*   **`Route`**: Represents an API endpoint, containing the path, method, and handler.
*   **`Group`**: Allows hierarchical organization of routes.
*   **`Binder`**: Handles the conversion between the HTTP request/response and Go structs.
*   **`Adapter`**: Bridges MiniAPI routes with an underlying HTTP router (like Echo, Gin, or standard `net/http`).
*   **`OpenAPI`**: Handles the generation and serving of OpenAPI specs.

## License

MIT
