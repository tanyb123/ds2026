#!/usr/bin/env python3
"""
RPC File Transfer Server using gRPC
Practical Work 2
"""

import grpc
from concurrent import futures
import filetransfer_pb2
import filetransfer_pb2_grpc
import os
import argparse

STORAGE_DIR = "./storage"

class FileTransferService(filetransfer_pb2_grpc.FileTransferServiceServicer):
    def __init__(self, storage_dir):
        self.storage_dir = storage_dir
        os.makedirs(storage_dir, exist_ok=True)
    
    def SendFile(self, request_iterator, context):
        """Receive file from client (streaming)"""
        filename = None
        filepath = None
        total_received = 0
        
        for chunk in request_iterator:
            if filename is None:
                filename = chunk.filename
                filepath = os.path.join(self.storage_dir, os.path.basename(filename))
                print(f"Receiving file: {filename}")
            
            # Write chunk to file
            with open(filepath, 'ab') as f:
                f.write(chunk.data)
                total_received += len(chunk.data)
            
            if chunk.is_last:
                print(f"File received: {filename} ({total_received} bytes)")
                return filetransfer_pb2.FileResponse(
                    success=True,
                    message=f"File {filename} received successfully",
                    file_size=total_received
                )
        
        return filetransfer_pb2.FileResponse(
            success=False,
            message="No data received"
        )
    
    def ReceiveFile(self, request, context):
        """Send file to client (streaming)"""
        filename = os.path.basename(request.filename)
        filepath = os.path.join(self.storage_dir, filename)
        
        if not os.path.exists(filepath):
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(f"File not found: {filename}")
            return
        
        file_size = os.path.getsize(filepath)
        print(f"Sending file: {filename} ({file_size} bytes)")
        
        with open(filepath, 'rb') as f:
            offset = 0
            while True:
                data = f.read(4096)
                if not data:
                    break
                
                yield filetransfer_pb2.FileChunk(
                    filename=filename,
                    data=data,
                    offset=offset,
                    is_last=(offset + len(data) >= file_size),
                    total_size=file_size
                )
                offset += len(data)
    
    def ListFiles(self, request, context):
        """List all files"""
        files = []
        for filename in os.listdir(self.storage_dir):
            filepath = os.path.join(self.storage_dir, filename)
            if os.path.isfile(filepath):
                stat = os.stat(filepath)
                files.append(filetransfer_pb2.FileInfo(
                    filename=filename,
                    size=stat.st_size,
                    modified_time=int(stat.st_mtime)
                ))
        
        return filetransfer_pb2.FileList(files=files)
    
    def DeleteFile(self, request, context):
        """Delete a file"""
        filename = os.path.basename(request.filename)
        filepath = os.path.join(self.storage_dir, filename)
        
        if os.path.exists(filepath):
            os.remove(filepath)
            print(f"File deleted: {filename}")
            return filetransfer_pb2.FileResponse(
                success=True,
                message=f"File {filename} deleted successfully"
            )
        else:
            return filetransfer_pb2.FileResponse(
                success=False,
                message=f"File not found: {filename}"
            )

def serve(port, storage_dir):
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    filetransfer_pb2_grpc.add_FileTransferServiceServicer_to_server(
        FileTransferService(storage_dir), server
    )
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    print(f"RPC File Transfer Server listening on port {port}")
    print(f"Storage directory: {storage_dir}")
    server.wait_for_termination()

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='RPC File Transfer Server')
    parser.add_argument('--port', type=int, default=50051, help='Server port')
    parser.add_argument('--dir', type=str, default=STORAGE_DIR, help='Storage directory')
    args = parser.parse_args()
    
    serve(args.port, args.dir)

