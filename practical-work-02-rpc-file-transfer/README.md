# Practical Work 2: RPC File Transfer

## Mô tả
Nâng cấp TCP File Transfer sang RPC sử dụng gRPC.

## Cài đặt
```bash
pip install grpcio grpcio-tools
```

## Generate gRPC code
```bash
python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. filetransfer.proto
```

## Sử dụng

### Server
```bash
python3 server.py --port 50051
```

### Client
```bash
# Send file
python3 client.py --server localhost:50051 --send-file test.txt

# Receive file
python3 client.py --server localhost:50051 --receive-file test.txt

# List files
python3 client.py --server localhost:50051 --list

# Delete file
python3 client.py --server localhost:50051 --delete test.txt
```
