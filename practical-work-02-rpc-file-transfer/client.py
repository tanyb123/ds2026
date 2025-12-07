#!/usr/bin/env python3
"""
RPC File Transfer Client using gRPC
Practical Work 2
"""

import grpc
import filetransfer_pb2
import filetransfer_pb2_grpc
import argparse
import os

def send_file(stub, filepath):
    """Send file to server"""
    if not os.path.exists(filepath):
        print(f"Error: File not found: {filepath}")
        return
    
    filename = os.path.basename(filepath)
    file_size = os.path.getsize(filepath)
    
    print(f"Sending file: {filename} ({file_size} bytes)")
    
    def file_chunks():
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
    
    response = stub.SendFile(file_chunks())
    if response.success:
        print(f"File sent successfully: {response.message}")
    else:
        print(f"Error: {response.message}")

def receive_file(stub, filename, output_path=None):
    """Receive file from server"""
    if output_path is None:
        output_path = filename
    
    print(f"Receiving file: {filename}")
    
    request = filetransfer_pb2.FileRequest(filename=filename)
    total_received = 0
    
    with open(output_path, 'wb') as f:
        for chunk in stub.ReceiveFile(request):
            f.write(chunk.data)
            total_received += len(chunk.data)
            if chunk.is_last:
                break
    
    print(f"File received: {output_path} ({total_received} bytes)")

def list_files(stub):
    """List files on server"""
    request = filetransfer_pb2.Empty()
    file_list = stub.ListFiles(request)
    
    print("Files on server:")
    print("=" * 50)
    for file_info in file_list.files:
        print(f"  {file_info.filename} - {file_info.size} bytes")

def delete_file(stub, filename):
    """Delete file on server"""
    request = filetransfer_pb2.FileRequest(filename=filename)
    response = stub.DeleteFile(request)
    
    if response.success:
        print(f"File deleted: {response.message}")
    else:
        print(f"Error: {response.message}")

def main():
    parser = argparse.ArgumentParser(description='RPC File Transfer Client')
    parser.add_argument('--server', type=str, default='localhost:50051', help='Server address')
    parser.add_argument('--send-file', type=str, help='File to send')
    parser.add_argument('--receive-file', type=str, help='File to receive')
    parser.add_argument('--list', action='store_true', help='List files')
    parser.add_argument('--delete', type=str, help='File to delete')
    args = parser.parse_args()
    
    with grpc.insecure_channel(args.server) as channel:
        stub = filetransfer_pb2_grpc.FileTransferServiceStub(channel)
        
        if args.send_file:
            send_file(stub, args.send_file)
        elif args.receive_file:
            receive_file(stub, args.receive_file)
        elif args.list:
            list_files(stub)
        elif args.delete:
            delete_file(stub, args.delete)
        else:
            print("Usage:")
            print("  --send-file <file>    Send file")
            print("  --receive-file <file> Receive file")
            print("  --list                List files")
            print("  --delete <file>       Delete file")

if __name__ == '__main__':
    main()

