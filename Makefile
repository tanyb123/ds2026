.PHONY: proto build server client clean

# Generate gRPC code from proto files
proto:
	@echo "Generating gRPC code..."
	@mkdir -p proto/generated
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/shell.proto

# Build all binaries
build: proto
	@echo "Building binaries..."
	@mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

# Build server only
server: proto
	@mkdir -p bin
	go build -o bin/server ./cmd/server

# Build client only
client: proto
	@mkdir -p bin
	go build -o bin/client ./cmd/client

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin
	rm -rf proto/generated
	find . -name "*.pb.go" -delete

# Run server
run-server: server
	./bin/server

# Run client (example)
run-client: client
	./bin/client --server localhost:50051 --command "ls -la"

# Install dependencies
deps:
	go mod download
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

