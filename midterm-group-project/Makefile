.PHONY: build server client admin clean run-server run-client run-admin

# OS-aware binary extension
EXT :=
ifeq ($(OS),Windows_NT)
EXT := .exe
endif

BIN := bin

# Build all binaries
build: server client admin

# Build server
server:
	@echo "Building server..."
	@mkdir -p $(BIN)
	go build -o $(BIN)/server$(EXT) ./server

# Build client
client:
	@echo "Building client..."
	@mkdir -p $(BIN)
	go build -o $(BIN)/client$(EXT) ./client

# Build admin tool
admin:
	@echo "Building admin tool..."
	@mkdir -p $(BIN)
	go build -o $(BIN)/admin$(EXT) ./admin

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BIN)/

# Run server
run-server: server
	@echo "Starting server..."
	./$(BIN)/server$(EXT)

# Run client (interactive)
run-client: client
	@echo "Starting client..."
	./$(BIN)/client$(EXT)

# Run admin
run-admin: admin
	@echo "Listing clients..."
	./$(BIN)/admin$(EXT)