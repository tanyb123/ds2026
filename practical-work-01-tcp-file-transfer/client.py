#!/usr/bin/env python3
"""
TCP File Transfer Client
Practical Work 1
"""

import socket
import sys
import os
import argparse

BUFFER_SIZE = 4096

def send_file(server_addr, port, filepath):
    """Send file to server"""
    if not os.path.exists(filepath):
        print(f"Error: File not found: {filepath}")
        return False
    
    filename = os.path.basename(filepath)
    file_size = os.path.getsize(filepath)
    
    print(f"Connecting to {server_addr}:{port}...")
    
    try:
        # Connect to server
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.connect((server_addr, port))
        
        # Send file info
        file_info = f"{filename}|{file_size}"
        sock.send(file_info.encode('utf-8'))
        
        # Wait for ACK
        response = sock.recv(1024).decode('utf-8')
        if response != "OK":
            print(f"Error: Server rejected: {response}")
            sock.close()
            return False
        
        print(f"Sending file: {filename} ({file_size} bytes)")
        
        # Send file data
        with open(filepath, 'rb') as f:
            total_sent = 0
            while total_sent < file_size:
                data = f.read(BUFFER_SIZE)
                if not data:
                    break
                sock.send(data)
                total_sent += len(data)
        
        # Wait for confirmation
        response = sock.recv(1024).decode('utf-8')
        if response == "DONE":
            print(f"File sent successfully!")
            return True
        else:
            print(f"Error: {response}")
            return False
    
    except Exception as e:
        print(f"Error: {e}")
        return False
    finally:
        sock.close()

def main():
    parser = argparse.ArgumentParser(description='TCP File Transfer Client')
    parser.add_argument('--server', type=str, default='localhost', help='Server address')
    parser.add_argument('--port', type=int, default=8080, help='Server port')
    parser.add_argument('--send-file', type=str, required=True, help='File to send')
    args = parser.parse_args()
    
    send_file(args.server, args.port, args.send_file)

if __name__ == '__main__':
    main()

