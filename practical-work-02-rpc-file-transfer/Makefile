.PHONY: proto build clean run-server run-client

# Generate gRPC code
proto:
	@echo "Generating gRPC code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/filetransfer.proto

# Build binaries
build: proto
	@echo "Building RPC File Transfer..."
	@mkdir -p bin
	go build -o bin/server ./server
	go build -o bin/client ./client

# Clean build artifacts
clean:
	rm -rf bin
	rm -rf storage
	find . -name "*.pb.go" -delete

# Run server
run-server: build
	./bin/server --port 50051

# Run client (send file)
run-client-send: build
	./bin/client --server localhost:50051 --send-file $(FILE)

# Run client (receive file)
run-client-receive: build
	./bin/client --server localhost:50051 --receive-file $(FILE)

# Run client (list files)
run-client-list: build
	./bin/client --server localhost:50051 --list

