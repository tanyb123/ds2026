# Practical Work 1: TCP File Transfer

## Mô tả
Hệ thống chuyển file 1-1 qua TCP/IP trong CLI.

## Cấu trúc
```
practical-work-01-tcp-file-transfer/
├── server.py      # TCP server
├── client.py      # TCP client
└── README.md
```

## Protocol
1. Client gửi: `filename|filesize`
2. Server trả: `OK` hoặc `ERROR|message`
3. Client gửi file data
4. Server trả: `DONE` hoặc `ERROR|message`

## Sử dụng

### Chạy Server
```bash
python3 server.py --port 8080
```

### Chạy Client
```bash
python3 client.py --server localhost --port 8080 --send-file test.txt
```
