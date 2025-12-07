CC = mpicc
CFLAGS = -Wall -Wextra -O2
TARGET = bin/mpi_file_transfer
SRC = src/mpi_file_transfer.c

.PHONY: build clean run

# Build
build:
	@echo "Building MPI File Transfer..."
	@mkdir -p bin
	$(CC) $(CFLAGS) -o $(TARGET) $(SRC)

# Clean
clean:
	rm -rf bin
	rm -f worker_*_*

# Run with 4 processes
run: build
	mpirun -np 4 $(TARGET) --send-file test.txt

# Run with custom number of processes
run-np:
	mpirun -np $(NP) $(TARGET) --send-file $(FILE)

# Test
test: build
	@echo "Creating test file..."
	@dd if=/dev/urandom of=test.txt bs=1024 count=100 2>/dev/null || \
	    python3 -c "import os; f=open('test.txt','wb'); f.write(os.urandom(102400)); f.close()"
	@echo "Running MPI file transfer with 4 processes..."
	mpirun -np 4 $(TARGET) --send-file test.txt
	@echo "Done!"

