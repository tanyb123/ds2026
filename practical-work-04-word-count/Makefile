CC = gcc
CFLAGS = -Wall -Wextra -O2 -pthread
TARGET = bin/wordcount
SRCS = src/main.c src/mapreduce.c

.PHONY: build clean test

# Build
build:
	@echo "Building Word Count MapReduce..."
	@mkdir -p bin
	$(CC) $(CFLAGS) -o $(TARGET) $(SRCS)

# Clean
clean:
	rm -rf bin
	rm -rf output/*

# Test
test: build
	@echo "Creating test input files..."
	@mkdir -p input
	@echo "the quick brown fox jumps over the lazy dog" > input/file1.txt
	@echo "the fox is quick and the dog is lazy" > input/file2.txt
	@echo "brown fox lazy dog" > input/file3.txt
	@echo "Running word count..."
	./bin/wordcount input/ output/
	@echo "Results:"
	@cat output/*

