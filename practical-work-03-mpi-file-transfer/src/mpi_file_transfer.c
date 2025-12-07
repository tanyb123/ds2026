#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <mpi.h>
#include "protocol.h"

#define COORDINATOR_RANK 0

// Function declarations
void coordinator_send_file(const char *filename, int num_workers);
void worker_receive_file(int rank);
void coordinator_receive_results(int num_workers, const char *output_file);

int main(int argc, char **argv) {
    int rank, size;
    int provided;
    
    // Initialize MPI with thread support
    MPI_Init_thread(&argc, &argv, MPI_THREAD_SINGLE, &provided);
    MPI_Comm_rank(MPI_COMM_WORLD, &rank);
    MPI_Comm_size(MPI_COMM_WORLD, &size);
    
    if (size < 2) {
        if (rank == 0) {
            fprintf(stderr, "Error: Need at least 2 processes (1 coordinator + 1 worker)\n");
            fprintf(stderr, "Usage: mpirun -np <N> %s --send-file <file>\n", argv[0]);
        }
        MPI_Finalize();
        return 1;
    }
    
    int num_workers = size - 1;
    
    // Parse arguments
    const char *filename = NULL;
    for (int i = 1; i < argc; i++) {
        if (strcmp(argv[i], "--send-file") == 0 && i + 1 < argc) {
            filename = argv[i + 1];
            break;
        }
    }
    
    if (rank == COORDINATOR_RANK) {
        // Coordinator process
        if (filename == NULL) {
            fprintf(stderr, "Error: --send-file <file> is required\n");
            MPI_Finalize();
            return 1;
        }
        
        printf("Coordinator (Rank %d): Starting file transfer\n", rank);
        printf("Number of workers: %d\n", num_workers);
        printf("File: %s\n", filename);
        
        coordinator_send_file(filename, num_workers);
        
        // Receive results from workers
        char output_file[256];
        snprintf(output_file, sizeof(output_file), "%s.received", filename);
        coordinator_receive_results(num_workers, output_file);
        
        printf("Coordinator: File transfer completed\n");
    } else {
        // Worker process
        printf("Worker (Rank %d): Ready to receive\n", rank);
        worker_receive_file(rank);
        printf("Worker (Rank %d): Finished\n", rank);
    }
    
    MPI_Finalize();
    return 0;
}

void coordinator_send_file(const char *filename, int num_workers) {
    FILE *file = fopen(filename, "rb");
    if (file == NULL) {
        fprintf(stderr, "Error: Cannot open file %s\n", filename);
        return;
    }
    
    // Get file size
    fseek(file, 0, SEEK_END);
    int64_t file_size = ftell(file);
    fseek(file, 0, SEEK_SET);
    
    // Prepare file info
    FileInfo info;
    strncpy(info.filename, filename, sizeof(info.filename) - 1);
    info.filename[sizeof(info.filename) - 1] = '\0';
    info.file_size = file_size;
    info.total_chunks = (file_size + CHUNK_SIZE - 1) / CHUNK_SIZE;
    
    printf("Coordinator: File size: %ld bytes, Chunks: %d\n", 
           (long)file_size, info.total_chunks);
    
    // Send file info to all workers
    for (int worker = 1; worker <= num_workers; worker++) {
        MPI_Send(&info, sizeof(FileInfo), MPI_BYTE, worker, TAG_FILE_INFO, MPI_COMM_WORLD);
    }
    
    // Distribute chunks to workers
    FileChunk chunk;
    int chunk_id = 0;
    int current_worker = 1;
    
    while (1) {
        chunk.chunk_id = chunk_id;
        chunk.chunk_size = fread(chunk.data, 1, CHUNK_SIZE, file);
        
        if (chunk.chunk_size == 0) {
            break;
        }
        
        // Send chunk to current worker (round-robin)
        MPI_Send(&chunk, sizeof(FileChunk), MPI_BYTE, current_worker, 
                 TAG_FILE_CHUNK, MPI_COMM_WORLD);
        
        printf("Coordinator: Sent chunk %d (%d bytes) to worker %d\n", 
               chunk_id, chunk.chunk_size, current_worker);
        
        chunk_id++;
        current_worker = (current_worker % num_workers) + 1;
    }
    
    // Send end signal to all workers
    for (int worker = 1; worker <= num_workers; worker++) {
        MPI_Send(NULL, 0, MPI_BYTE, worker, TAG_FILE_END, MPI_COMM_WORLD);
    }
    
    fclose(file);
    printf("Coordinator: All chunks sent\n");
}

void worker_receive_file(int rank) {
    FileInfo info;
    MPI_Status status;
    
    // Receive file info
    MPI_Recv(&info, sizeof(FileInfo), MPI_BYTE, COORDINATOR_RANK, 
             TAG_FILE_INFO, MPI_COMM_WORLD, &status);
    
    printf("Worker %d: Receiving file: %s (%ld bytes, %d chunks)\n", 
           rank, info.filename, (long)info.file_size, info.total_chunks);
    
    // Create output filename
    char output_filename[256];
    snprintf(output_filename, sizeof(output_filename), 
             "worker_%d_%s", rank, info.filename);
    
    FILE *output_file = fopen(output_filename, "wb");
    if (output_file == NULL) {
        fprintf(stderr, "Worker %d: Cannot create output file\n", rank);
        return;
    }
    
    int chunks_received = 0;
    int64_t total_received = 0;
    
    // Receive chunks
    while (1) {
        MPI_Probe(COORDINATOR_RANK, MPI_ANY_TAG, MPI_COMM_WORLD, &status);
        
        if (status.MPI_TAG == TAG_FILE_END) {
            // Receive and discard end message
            MPI_Recv(NULL, 0, MPI_BYTE, COORDINATOR_RANK, TAG_FILE_END, 
                     MPI_COMM_WORLD, &status);
            break;
        }
        
        if (status.MPI_TAG == TAG_FILE_CHUNK) {
            FileChunk chunk;
            MPI_Recv(&chunk, sizeof(FileChunk), MPI_BYTE, COORDINATOR_RANK, 
                     TAG_FILE_CHUNK, MPI_COMM_WORLD, &status);
            
            // Write chunk to file
            fwrite(chunk.data, 1, chunk.chunk_size, output_file);
            chunks_received++;
            total_received += chunk.chunk_size;
            
            printf("Worker %d: Received chunk %d (%d bytes), total: %ld bytes\n", 
                   rank, chunk.chunk_id, chunk.chunk_size, (long)total_received);
            
            // Send result back to coordinator
            ChunkResult result;
            result.chunk_id = chunk.chunk_id;
            result.status = 0;  // Success
            result.processed_size = chunk.chunk_size;
            
            MPI_Send(&result, sizeof(ChunkResult), MPI_BYTE, COORDINATOR_RANK, 
                     TAG_RESULT, MPI_COMM_WORLD);
        }
    }
    
    fclose(output_file);
    printf("Worker %d: Received %d chunks, total %ld bytes\n", 
           rank, chunks_received, (long)total_received);
}

void coordinator_receive_results(int num_workers, const char *output_file) {
    FILE *combined_file = fopen(output_file, "wb");
    if (combined_file == NULL) {
        fprintf(stderr, "Coordinator: Cannot create combined output file\n");
        return;
    }
    
    // Read and combine files from workers
    for (int worker = 1; worker <= num_workers; worker++) {
        char worker_file[256];
        snprintf(worker_file, sizeof(worker_file), "worker_%d_", worker);
        
        // Find the actual worker file (simplified - in real implementation,
        // we'd track which chunks went to which worker)
        // For now, just collect results
    }
    
    // Receive all results
    int total_chunks = 0;
    int64_t total_processed = 0;
    
    for (int i = 0; i < num_workers * 10; i++) {  // Estimate max chunks
        MPI_Status status;
        MPI_Probe(MPI_ANY_SOURCE, TAG_RESULT, MPI_COMM_WORLD, &status);
        
        ChunkResult result;
        MPI_Recv(&result, sizeof(ChunkResult), MPI_BYTE, MPI_ANY_SOURCE, 
                 TAG_RESULT, MPI_COMM_WORLD, &status);
        
        if (result.status == 0) {
            total_chunks++;
            total_processed += result.processed_size;
        }
    }
    
    fclose(combined_file);
    printf("Coordinator: Received results from workers\n");
    printf("Coordinator: Total chunks processed: %d, Total bytes: %ld\n", 
           total_chunks, (long)total_processed);
}

