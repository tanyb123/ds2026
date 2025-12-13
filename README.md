# Remote Shell RPC System

Hệ thống Remote Shell sử dụng RPC (Remote Procedure Call) trong Go, mô phỏng `kubectl exec` trên Kubernetes. Hệ thống hỗ trợ nhiều clients kết nối đồng thời đến một RPC server để thực thi các lệnh shell từ xa.

## Tính năng

- ✅ RPC Server xử lý nhiều clients đồng thời
- ✅ Remote shell command execution
- ✅ Session management cho mỗi client
- ✅ Environment variables per session
- ✅ Working directory per session
- ✅ Interactive và non-interactive mode
- ✅ Client tracking và listing

## Kiến trúc

```
┌─────────┐     ┌─────────┐     ┌─────────┐
│ Client1 │────▶│         │◀────│ Client2 │
└─────────┘     │  RPC    │     └─────────┘
                │ Server  │
┌─────────┐     │         │     ┌─────────┐
│ Client3 │────▶│         │◀────│ Client4 │
└─────────┘     └─────────┘     └─────────┘
```

## Cài đặt

### Yêu cầu
- Go 1.21 hoặc cao hơn

### Cài đặt Go (Nếu chưa có)

1. Tải Go từ: https://go.dev/dl/
2. Chọn bản Windows (ví dụ: `go1.21.x.windows-amd64.msi`)
3. Chạy installer và làm theo hướng dẫn
4. **Quan trọng**: Sau khi cài đặt, **khởi động lại PowerShell**

Kiểm tra cài đặt:
```powershell
go version
```

### Build

#### Windows (PowerShell) - Khuyên dùng
```powershell
# Dùng script tự động
.\build.ps1

# Hoặc build thủ công
go build -o bin\server.exe ./server
go build -o bin\client.exe ./client
go build -o bin\admin.exe ./admin
```

#### Linux/Mac
```bash
# Build server
go build -o bin/server ./server

# Build client
go build -o bin/client ./client

# Build admin tool
go build -o bin/admin ./admin
```

> 💡 **Lưu ý**: Nếu gặp lỗi "go is not recognized", xem hướng dẫn chi tiết trong [SETUP.md](SETUP.md)

## Sử dụng

### 1. Khởi động Server

**Windows (PowerShell):**
```powershell
# Cách 1: Dùng script (tự động build nếu cần)
.\run-server.ps1

# Cách 2: Chạy trực tiếp
.\bin\server.exe
```

**Linux/Mac:**
```bash
./bin/server
```

Server sẽ chạy trên port `8080` và chờ các clients kết nối.

### 2. Chạy Client (Interactive Mode)

**Windows (PowerShell):**
```powershell
# Cách 1: Dùng script
.\run-client.ps1 -ClientID my-client-1

# Cách 2: Chạy trực tiếp
.\bin\client.exe -id my-client-1
```

**Linux/Mac:**
```bash
./bin/client -id my-client-1
```

Hoặc để tự động generate client ID:
```powershell
# Windows
.\bin\client.exe

# Linux/Mac
./bin/client
```

### 3. Chạy Client (Non-Interactive Mode)

Thực thi một lệnh và thoát:

**Windows:**
```powershell
.\bin\client.exe -id my-client-1 -cmd "dir"
.\bin\client.exe -id my-client-1 -cmd "echo Hello World"
```

**Linux/Mac:**
```bash
./bin/client -id my-client-1 -cmd "ls -la"
./bin/client -id my-client-1 -cmd "echo Hello World"
```

### 4. Admin Tool - Liệt kê Clients

```powershell
# Windows
.\bin\admin.exe

# Linux/Mac
./bin/admin
```

## Ví dụ sử dụng

### Terminal 1: Server
```bash
$ ./bin/server
Remote Shell RPC Server started on :8080
Waiting for clients...
```

### Terminal 2: Client 1
```bash
$ ./bin/client -id client1
Connected to server localhost:8080 as client1
[client1@remote]$ pwd
/home/user
[client1@remote]$ ls -la
total 24
drwxr-xr-x  3 user user 4096 Jan 15 10:00 .
...
[client1@remote]$ cd /tmp
Directory changed
[client1@remote]$ setenv MY_VAR hello
Set MY_VAR=hello
[client1@remote]$ echo $MY_VAR
hello
[client1@remote]$ exit
```

### Terminal 3: Client 2
```bash
$ ./bin/client -id client2
Connected to server localhost:8080 as client2
[client2@remote]$ pwd
/home/user
[client2@remote]$ cd /var/log
Directory changed
[client2@remote]$ ls
...
```

### Terminal 4: Admin
```bash
$ ./bin/admin
Active clients (2):
  1. client1
  2. client2
```

## RPC Methods

### RemoteShellService.Execute
Thực thi một lệnh shell.

**Request:**
```go
type CommandRequest struct {
    Command string
    Args    []string
    ID      string
}
```

**Response:**
```go
type CommandResponse struct {
    Output   string
    Error    string
    ExitCode int
    ID       string
}
```

### RemoteShellService.SetEnv
Thiết lập biến môi trường cho session.

### RemoteShellService.ChangeDir
Thay đổi thư mục làm việc cho session.

### RemoteShellService.ListClients
Liệt kê tất cả clients đang active.

## So sánh với kubectl exec

| Tính năng | kubectl exec | Remote Shell RPC |
|-----------|--------------|------------------|
| Remote execution | ✅ | ✅ |
| Multiple clients | ✅ (multiple pods) | ✅ |
| Session management | ✅ (per pod) | ✅ (per client ID) |
| Environment vars | ✅ | ✅ |
| Working directory | ✅ | ✅ |
| Streaming output | ✅ | ⚠️ (batch) |
| Interactive TTY | ✅ | ⚠️ (basic) |

## Cải tiến có thể thêm

- [ ] Streaming output (real-time)
- [ ] Interactive TTY support
- [ ] Authentication và authorization
- [ ] TLS/SSL encryption
- [ ] Command history
- [ ] File transfer (scp-like)
- [ ] gRPC thay vì net/rpc
- [ ] Metrics và monitoring

## Troubleshooting

### Lỗi: "go is not recognized"
- **Nguyên nhân**: Go chưa được cài đặt hoặc chưa có trong PATH
- **Giải pháp**: 
  1. Cài đặt Go từ https://go.dev/dl/
  2. Khởi động lại PowerShell
  3. Kiểm tra lại: `go version`

### Lỗi: "Access denied" khi chạy PowerShell script
- **Giải pháp**: Chạy lệnh này một lần:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

Xem thêm hướng dẫn chi tiết trong [SETUP.md](SETUP.md)

## License

MIT

