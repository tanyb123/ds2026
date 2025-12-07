#ifndef PROTOCOL_H
#define PROTOCOL_H

#include <mpi.h>
#include <stdint.h>

// Message tags
#define TAG_FILE_INFO    1
#define TAG_FILE_CHUNK   2
#define TAG_FILE_END     3
#define TAG_RESULT       4
#define TAG_REQUEST      5

// Chunk size (4KB)
#define CHUNK_SIZE 4096

// File info structure
typedef struct {
    char filename[256];
    int64_t file_size;
    int total_chunks;
} FileInfo;

// Chunk structure
typedef struct {
    int chunk_id;
    int chunk_size;
    char data[CHUNK_SIZE];
} FileChunk;

// Result structure
typedef struct {
    int chunk_id;
    int status;  // 0 = success, -1 = error
    int64_t processed_size;
} ChunkResult;

#endif // PROTOCOL_H

