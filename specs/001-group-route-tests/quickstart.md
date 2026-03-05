# Quickstart: Testing Group & Route Packages

To run the updated and new automated tests for the `group` and `route` packages:

1. Ensure you are in the project root directory.
2. Run the `Taskfile` test command or standard `go test` targeting the specific packages:

```bash
# Run tests for both packages
go test -v ./group/... ./route/...

# Run tests with coverage
task test-cover
```

To regenerate any necessary mocks during development:

```bash
task mock
```