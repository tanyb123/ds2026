# USTH 2026: DISTRIBUTED SYSTEM

Students are expected to:
* Fork this repository to your github account.
* Push your commits regularly, with **PROPER** commit **MESSAGE**.

## Student Info
* Student Name: Nguyễn Duy Tân
* Student ID: 22BA13278
* Student Group ID: Group 14

## Practical Works

### Practical Work 1: TCP File Transfer
- **Location**: `practical-work-01-tcp-file-transfer/`
- **Tech**: Python, TCP sockets
- **Files**: `server.py`, `client.py`

### Practical Work 2: RPC File Transfer
- **Location**: `practical-work-02-rpc-file-transfer/`
- **Tech**: Python, gRPC
- **Files**: `server.py`, `client.py`, `filetransfer.proto`

### Practical Work 3: MPI File Transfer
- **Location**: `practical-work-03-mpi-file-transfer/`
- **Tech**: Python, mpi4py
- **Files**: `mpi_file_transfer.py`

### Practical Work 4: Word Count (MapReduce)
- **Location**: `practical-work-04-word-count/`
- **Tech**: Python, multiprocessing
- **Files**: `wordcount.py`

## Quick Start

### Prerequisites
```bash
pip install grpcio grpcio-tools mpi4py
```

### Run Examples

**TCP File Transfer:**
```bash
# Terminal 1
cd practical-work-01-tcp-file-transfer
python3 server.py

# Terminal 2
python3 client.py --send-file test.txt
```

**RPC File Transfer:**
```bash
# Generate gRPC code first
cd practical-work-02-rpc-file-transfer
python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. filetransfer.proto

# Terminal 1
python3 server.py

# Terminal 2
python3 client.py --send-file test.txt
```

**MPI File Transfer:**
```bash
cd practical-work-03-mpi-file-transfer
mpirun -np 4 python3 mpi_file_transfer.py --send-file test.txt
```

**Word Count:**
```bash
cd practical-work-04-word-count
mkdir -p input output
echo "hello world" > input/file1.txt
python3 wordcount.py input/ output/
```
