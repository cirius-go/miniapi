# Data Model: Binder Test Fixtures

As a testing feature, this plan does not introduce new production data models. It introduces specific struct fixtures within the test suite to validate binding behaviors.

### Test Fixtures

- **PrimitiveStruct**: Contains fields of various Go primitive types (string, int, float, bool) mapped via `path`, `query`, and `header` tags.
- **SliceStruct**: Contains fields of slice types mapped via `query` and `header` tags.
- **PointerStruct**: Contains fields defined as pointers to primitives and slices to test nil-initialization and allocation.
- **JSONBodyStruct**: A standard struct to test JSON unmarshaling via the `Body` struct tag detection.
- **RawBodyStruct**: Structs containing `io.Reader` and `[]byte` field definitions for raw stream testing.

These fixtures will reside within the `binder_test.go` file.