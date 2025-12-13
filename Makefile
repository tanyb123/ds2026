.PHONY: build server client admin clean run-server run-client run-admin

# Build all binaries
build: server client admin

# Build server
server:
	@echo "Building server..."
	@mkdir -p bin
	go build -o bin/server.exe ./server

# Build client
client:
	@echo "Building client..."
	@mkdir -p bin
	go build -o bin/client.exe ./client

# Build admin tool
admin:
	@echo "Building admin tool..."
	@mkdir -p bin
	go build -o bin/admin.exe ./admin

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/

# Run server
run-server: server
	@echo "Starting server..."
	./bin/server.exe

# Run client (interactive)
run-client: client
	@echo "Starting client..."
	./bin/client.exe

# Run admin
run-admin: admin
	@echo "Listing clients..."
	./bin/admin.exe




