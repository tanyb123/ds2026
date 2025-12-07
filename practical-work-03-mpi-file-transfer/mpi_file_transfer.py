#!/usr/bin/env python3
"""
MPI File Transfer
Practical Work 3
"""

from mpi4py import MPI
import sys
import os
import argparse

CHUNK_SIZE = 4096

def coordinator_send_file(filename, num_workers):
    """Coordinator sends file to workers"""
    if not os.path.exists(filename):
        print(f"Error: File not found: {filename}")
        return
    
    file_size = os.path.getsize(filename)
    print(f"Coordinator: File size: {file_size} bytes")
    
    # Send file info to all workers
    file_info = {'filename': filename, 'size': file_size}
    for worker in range(1, num_workers + 1):
        MPI.COMM_WORLD.send(file_info, dest=worker, tag=1)
    
    # Distribute chunks to workers
    chunk_id = 0
    current_worker = 1
    
    with open(filename, 'rb') as f:
        while True:
            chunk_data = f.read(CHUNK_SIZE)
            if not chunk_data:
                break
            
            chunk = {'id': chunk_id, 'data': chunk_data}
            MPI.COMM_WORLD.send(chunk, dest=current_worker, tag=2)
            
            print(f"Coordinator: Sent chunk {chunk_id} to worker {current_worker}")
            chunk_id += 1
            current_worker = (current_worker % num_workers) + 1
    
    # Send end signal
    for worker in range(1, num_workers + 1):
        MPI.COMM_WORLD.send(None, dest=worker, tag=3)
    
    print(f"Coordinator: All chunks sent")

def worker_receive_file(rank):
    """Worker receives file chunks"""
    # Receive file info
    file_info = MPI.COMM_WORLD.recv(source=0, tag=1)
    filename = file_info['filename']
    file_size = file_info['size']
    
    print(f"Worker {rank}: Receiving file: {filename}")
    
    output_filename = f"worker_{rank}_{os.path.basename(filename)}"
    chunks_received = 0
    total_received = 0
    
    with open(output_filename, 'wb') as f:
        while True:
            status = MPI.Status()
            chunk = MPI.COMM_WORLD.recv(source=0, tag=MPI.ANY_TAG, status=status)
            
            if status.tag == 3:  # End signal
                break
            
            if status.tag == 2:  # File chunk
                f.write(chunk['data'])
                chunks_received += 1
                total_received += len(chunk['data'])
                print(f"Worker {rank}: Received chunk {chunk['id']}, total: {total_received} bytes")
    
    print(f"Worker {rank}: Received {chunks_received} chunks, {total_received} bytes")

def main():
    comm = MPI.COMM_WORLD
    rank = comm.Get_rank()
    size = comm.Get_size()
    
    if size < 2:
        if rank == 0:
            print("Error: Need at least 2 processes")
            print("Usage: mpirun -np <N> python3 mpi_file_transfer.py --send-file <file>")
        return
    
    num_workers = size - 1
    
    parser = argparse.ArgumentParser()
    parser.add_argument('--send-file', type=str, required=(rank == 0))
    args = parser.parse_args()
    
    if rank == 0:
        # Coordinator
        if not args.send_file:
            print("Error: --send-file is required")
            return
        coordinator_send_file(args.send_file, num_workers)
    else:
        # Worker
        worker_receive_file(rank)

if __name__ == '__main__':
    main()

