#!/usr/bin/env python3
"""
TCP File Transfer Server
Practical Work 1
"""

import socket
import sys
import os
import argparse

BUFFER_SIZE = 4096
RECEIVE_DIR = "./received"

def handle_client(conn, addr, receive_dir):
    """Handle client connection"""
    print(f"Connection from {addr}")
    
    try:
        # Receive file info (filename and size)
        file_info = conn.recv(1024).decode('utf-8')
        if not file_info:
            return
        
        parts = file_info.split('|')
        if len(parts) != 2:
            conn.send(b"ERROR|Invalid request format")
            return
        
        filename, file_size = parts[0], int(parts[1])
        print(f"Receiving file: {filename} ({file_size} bytes)")
        
        # Send ACK
        conn.send(b"OK")
        
        # Receive file data
        filepath = os.path.join(receive_dir, os.path.basename(filename))
        total_received = 0
        
        with open(filepath, 'wb') as f:
            while total_received < file_size:
                data = conn.recv(min(BUFFER_SIZE, file_size - total_received))
                if not data:
                    break
                f.write(data)
                total_received += len(data)
        
        if total_received == file_size:
            print(f"File received successfully: {filepath} ({total_received} bytes)")
            conn.send(b"DONE")
        else:
            print(f"Error: Incomplete transfer ({total_received}/{file_size} bytes)")
            os.remove(filepath)
            conn.send(b"ERROR|Incomplete transfer")
    
    except Exception as e:
        print(f"Error handling client: {e}")
        conn.send(f"ERROR|{str(e)}".encode('utf-8'))
    finally:
        conn.close()

def main():
    parser = argparse.ArgumentParser(description='TCP File Transfer Server')
    parser.add_argument('--port', type=int, default=8080, help='Server port')
    parser.add_argument('--dir', type=str, default=RECEIVE_DIR, help='Receive directory')
    args = parser.parse_args()
    
    # Create receive directory
    os.makedirs(args.dir, exist_ok=True)
    
    # Create socket
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server_socket.bind(('', args.port))
    server_socket.listen(5)
    
    print(f"TCP File Transfer Server listening on port {args.port}")
    print(f"Receiving files to: {args.dir}")
    
    try:
        while True:
            conn, addr = server_socket.accept()
            handle_client(conn, addr, args.dir)
    except KeyboardInterrupt:
        print("\nServer shutting down...")
    finally:
        server_socket.close()

if __name__ == '__main__':
    main()

