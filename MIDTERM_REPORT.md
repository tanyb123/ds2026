# [DS] Remote Shell RPC System - Midterm Report

**Project Title:** Remote Shell Execution System using RPC (Multiple Clients Support)

**Group ID:** [GID] - *[Điền Group ID của bạn]*

**Team Members:**
- [Member 1 Name] - [Role/Percentage]
- [Member 2 Name] - [Role/Percentage]
- [Member 3 Name] - [Role/Percentage]

**Date:** December 16, 2025

---

## 1. Description of Project

### 1.1 Overview
Remote Shell RPC System là một hệ thống phân tán cho phép nhiều clients kết nối đồng thời đến một RPC server để thực thi các lệnh shell từ xa. Hệ thống mô phỏng chức năng `kubectl exec` trong Kubernetes, cho phép quản lý và thực thi lệnh trên các máy tính từ xa thông qua giao thức RPC.

### 1.2 What it brings to end users

**Cho System Administrators:**
- Quản lý nhiều máy tính từ xa từ một điểm trung tâm
- Thực thi lệnh shell trên các máy remote mà không cần SSH trực tiếp
- Theo dõi và quản lý nhiều sessions đồng thời
- Kiểm soát môi trường làm việc (working directory, environment variables) cho mỗi client

**Cho Developers:**
- Remote debugging và testing trên các máy khác nhau
- Automation scripts có thể kết nối và thực thi lệnh từ xa
- Development environment management

**Cho Distributed Systems Learning:**
- Hiểu cách RPC hoạt động trong thực tế
- Học về concurrent programming với goroutines
- Session management trong distributed systems
- Network communication protocols

### 1.3 Key Features
- ✅ **Multiple Clients Support**: Hỗ trợ nhiều clients kết nối đồng thời
- ✅ **Session Management**: Mỗi client có session riêng với working directory và environment variables
- ✅ **Interactive & Non-Interactive Mode**: Hỗ trợ cả hai chế độ
- ✅ **Client Tracking**: Theo dõi và liệt kê tất cả clients đang active
- ✅ **Cross-platform**: Hoạt động trên Windows, Linux, và macOS
- ✅ **Concurrent Processing**: Xử lý nhiều requests đồng thời với goroutines

---

## 2. Architecture Design

### 2.1 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    RPC Server (Port 8080)                    │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         RemoteShellService (RPC Service)              │   │
│  │  ┌────────────────────────────────────────────────┐  │   │
│  │  │  Session Manager                                │  │   │
│  │  │  - Session Storage (map[string]*Session)        │  │   │
│  │  │  - Mutex for thread-safe operations             │  │   │
│  │  └────────────────────────────────────────────────┘  │   │
│  │                                                       │   │
│  │  RPC Methods:                                        │   │
│  │  - Register(clientID)                                │   │
│  │  - Execute(command)                                  │   │
│  │  - SetEnv(key, value)                                │   │
│  │  - ChangeDir(directory)                              │   │
│  │  - ListClients()                                     │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
         ▲                ▲                ▲
         │                │                │
         │                │                │
    ┌────┴────┐      ┌────┴────┐      ┌────┴────┐
    │ Client1 │      │ Client2 │      │ Client3 │
    │  (RPC   │      │  (RPC   │      │  (RPC   │
    │ Client) │      │ Client) │      │ Client) │
    └─────────┘      └─────────┘      └─────────┘
```

### 2.2 Component Description

#### 2.2.1 RPC Server
- **Location**: `server/main.go`
- **Responsibilities**:
  - Lắng nghe kết nối từ clients trên port 8080
  - Xử lý RPC requests từ nhiều clients đồng thời
  - Quản lý sessions cho mỗi client
  - Thực thi shell commands an toàn
  - Track và log tất cả operations

#### 2.2.2 RPC Client
- **Location**: `client/main.go`
- **Responsibilities**:
  - Kết nối đến RPC server
  - Gửi commands đến server
  - Hiển thị output từ server
  - Quản lý interactive session

#### 2.2.3 Admin Tool
- **Location**: `admin/main.go`
- **Responsibilities**:
  - Liệt kê tất cả clients đang active
  - Monitor system status

### 2.3 Multiple Computer Connection

Hệ thống hỗ trợ **nhiều computers kết nối** theo mô hình sau, **hỗ trợ cả mạng LAN và Internet**:

#### 2.3.1 Local Network (LAN) Connection

```
Computer A (Server)          Computer B (Client 1)        Computer C (Client 2)
192.168.1.100               192.168.1.101             192.168.1.102
┌──────────────┐             ┌──────────────┐             ┌──────────────┐
│              │             │              │             │              │
│  RPC Server  │◄────────────│ RPC Client   │             │              │
│  :8080       │             │              │             │              │
│              │             └──────────────┘             │              │
│              │                                          │              │
│              │◄─────────────────────────────────────────│ RPC Client   │
│              │                                          │              │
└──────────────┘                                          └──────────────┘
         │                          │                          │
         └──────────────────────────┴──────────────────────────┘
                            Router (192.168.1.1)
```

#### 2.3.2 Internet Connection (Different Networks)

```
Computer A (Server)          Computer B (Client 1)        Computer C (Client 2)
Public IP: 203.0.113.1      Public IP: 198.51.100.1   Public IP: 192.0.2.1
┌──────────────┐             ┌──────────────┐             ┌──────────────┐
│              │             │              │             │              │
│  RPC Server  │◄────────────│ RPC Client   │             │              │
│  :8080       │             │              │             │              │
│              │             └──────────────┘             │              │
│              │                                          │              │
│              │◄─────────────────────────────────────────│ RPC Client   │
│              │                                          │              │
└──────────────┘                                          └──────────────┘
         │                          │                          │
         └──────────────────────────┴──────────────────────────┘
                            Internet
```

**Network Topology:**
- **Star Topology**: Server ở trung tâm, tất cả clients kết nối đến server
- **TCP/IP Protocol**: Sử dụng TCP để đảm bảo reliable communication
- **Port**: 8080 (bind to 0.0.0.0:8080 để nhận kết nối từ mọi interface)
- **Network Support**: 
  - ✅ **Local Network (LAN)**: Clients cùng mạng với server
  - ✅ **Internet**: Clients khác mạng, kết nối qua public IP
  - ✅ **NAT/Port Forwarding**: Server sau NAT, clients kết nối qua router's public IP

**Connection Scenarios:**

**Scenario 1: Same LAN**
- Server: `192.168.1.100:8080`
- Client 1: `192.168.1.101` → Connect: `.\bin\client.exe -server 192.168.1.100:8080 -id client1`
- Client 2: `192.168.1.102` → Connect: `.\bin\client.exe -server 192.168.1.100:8080 -id client2`
- **Không cần VPN**, không cần public IP

**Scenario 2: Different Networks (Internet)**
- Server: `203.0.113.1:8080` (public IP)
- Client 1: `198.51.100.1` → Connect: `.\bin\client.exe -server 203.0.113.1:8080 -id client1`
- Client 2: `192.0.2.1` → Connect: `.\bin\client.exe -server 203.0.113.1:8080 -id client2`
- **Không cần VPN**, chỉ cần server có public IP và firewall mở port 8080

**Scenario 3: Server behind NAT**
- Server: `192.168.1.100` (private), Router: `203.0.113.1` (public)
- Router: Port forwarding 8080 → 192.168.1.100:8080
- Clients: Connect to `203.0.113.1:8080`
- **Không cần VPN**, chỉ cần port forwarding

**Connection Flow:**
1. Server khởi động và lắng nghe trên `0.0.0.0:8080` (tất cả network interfaces)
2. Server hiển thị local IP addresses để clients biết cách kết nối
3. Client 1 kết nối đến server qua TCP (LAN hoặc Internet)
4. Client 1 đăng ký với server (Register RPC call)
5. Client 2 kết nối đến server qua TCP (cùng lúc với Client 1, có thể từ mạng khác)
6. Client 2 đăng ký với server
7. Cả hai clients có thể thực thi commands đồng thời
8. Server xử lý requests từ cả hai clients song song (goroutines)

**Key Features:**
- ✅ **Multi-network support**: Clients có thể ở mạng khác nhau
- ✅ **No VPN required**: Không cần VPN nếu server có public IP
- ✅ **Flexible connection**: Hỗ trợ cả LAN và Internet
- ✅ **Firewall friendly**: Chỉ cần mở 1 port (8080)

### 2.4 Concurrency Model

Server sử dụng **goroutines** để xử lý nhiều clients đồng thời:

```go
for {
    conn, err := listener.Accept()
    go func(conn net.Conn) {
        rpc.ServeConn(conn)  // Mỗi client chạy trong goroutine riêng
    }(conn)
}
```

**Benefits:**
- Non-blocking: Một client không block các clients khác
- Scalable: Có thể xử lý hàng trăm clients đồng thời
- Efficient: Goroutines nhẹ hơn threads truyền thống

---

## 3. Protocol Design

### 3.1 RPC Protocol Overview

Hệ thống sử dụng **Go's built-in RPC (net/rpc)** protocol:
- **Transport**: TCP/IP
- **Encoding**: Go's gob encoding (binary)
- **Connection**: Persistent TCP connections

### 3.2 Communication Protocol

#### 3.2.1 Connection Establishment

```
Client                          Server
  │                               │
  │─── TCP Connect :8080 ────────▶│
  │                               │
  │─── RPC: Register(clientID) ──▶│
  │◀── Response: "Registered" ────│
  │                               │
  │─── Ready for commands ────────│
```

#### 3.2.2 Command Execution Protocol

```
Client                          Server
  │                               │
  │─── RPC: Execute(command) ────▶│
  │                               │─── Execute shell command
  │                               │─── Capture output
  │◀── Response: {output, exit} ──│
  │                               │
```

### 3.3 RPC Methods Specification

#### 3.3.1 Register
**Purpose**: Đăng ký client với server khi kết nối

**Request:**
```go
clientID string
```

**Response:**
```go
message string  // "Client {ID} registered successfully"
```

**Flow:**
1. Client kết nối TCP đến server
2. Client gọi `RemoteShellService.Register(clientID)`
3. Server tạo session mới cho client
4. Server trả về confirmation message

#### 3.3.2 Execute
**Purpose**: Thực thi shell command

**Request:**
```go
type CommandRequest struct {
    Command string   // Shell command to execute
    Args    []string // Command arguments (reserved for future)
    ID      string   // Client ID
}
```

**Response:**
```go
type CommandResponse struct {
    Output   string // Command output (stdout + stderr)
    Error    string // Error message if any
    ExitCode int    // Exit code (0 = success)
    ID       string // Client ID
}
```

**Flow:**
1. Client gửi command và client ID
2. Server tìm session của client
3. Server thực thi command trong session's working directory
4. Server capture output và exit code
5. Server trả về response

#### 3.3.3 SetEnv
**Purpose**: Thiết lập environment variable cho session

**Request:**
```go
map[string]string {
    "client_id": "client1",
    "key": "MY_VAR",
    "value": "hello"
}
```

**Response:**
```go
string // "Set MY_VAR=hello for client client1"
```

#### 3.3.4 ChangeDir
**Purpose**: Thay đổi working directory cho session

**Request:**
```go
map[string]string {
    "client_id": "client1",
    "dir": "/path/to/directory"
}
```

**Response:**
```go
string // "Changed directory to /path/to/directory for client client1"
```

#### 3.3.5 ListClients
**Purpose**: Liệt kê tất cả clients đang active

**Request:**
```go
string // Empty string (not used)
```

**Response:**
```go
[]string // ["client1", "client2", "client3"]
```

### 3.4 Message Flow Example

**Scenario**: Client 1 và Client 2 cùng thực thi commands

```
Time    Client 1                    Server                      Client 2
─────────────────────────────────────────────────────────────────────────
T0      Connect ────────────────────▶│
T1      Register("client1") ────────▶│
T2      │◀─── "Registered" ──────────│
T3      │                              │◀─────── Connect
T4      │                              │◀─────── Register("client2")
T5      │                              │───────▶ "Registered"
T6      Execute("pwd") ───────────────▶│
T7      │                              │─── Execute command
T8      │◀─── Response ────────────────│
T9      │                              │◀─────── Execute("ls")
T10     │                              │───────▶ Response
```

**Key Points:**
- Server xử lý requests từ nhiều clients **song song** (concurrent)
- Mỗi client có session riêng (working directory, env vars)
- Server sử dụng mutex để đảm bảo thread-safety

### 3.5 Error Handling

**Connection Errors:**
- Client tự động retry hoặc hiển thị error message
- Server log errors và continue serving other clients

**Command Execution Errors:**
- Exit code != 0 được trả về trong response
- Error message được capture và trả về

**Session Errors:**
- Nếu client chưa register, server tự động tạo session khi cần
- Invalid client ID được handle gracefully

---

## 4. Deployment Guideline

### 4.1 Prerequisites

**Required Software:**
- Go 1.21 or higher
- Git (for cloning repository)
- Terminal/PowerShell (Windows) or Terminal (Linux/Mac)

**System Requirements:**
- Minimum: 512MB RAM, 100MB disk space
- Network: TCP port 8080 available (or configurable)

### 4.2 Installation Steps

#### Step 1: Install Go

**Windows:**
1. Download Go from https://go.dev/dl/
2. Run installer (e.g., `go1.21.x.windows-amd64.msi`)
3. Restart PowerShell/Terminal
4. Verify: `go version`

**Linux:**
```bash
sudo apt-get update
sudo apt-get install golang-go
go version
```

**macOS:**
```bash
brew install go
go version
```

#### Step 2: Clone/Download Project

```bash
# If using Git
git clone <repository-url>
cd remote-shell-rpc

# Or download and extract ZIP file
```

#### Step 3: Build Project

**Windows (PowerShell):**
```powershell
# Option 1: Use build script
.\build.ps1

# Option 2: Manual build
go build -o bin\server.exe ./server
go build -o bin\client.exe ./client
go build -o bin\admin.exe ./admin
```

**Linux/Mac:**
```bash
# Option 1: Use Makefile
make build

# Option 2: Manual build
go build -o bin/server ./server
go build -o bin/client ./client
go build -o bin/admin ./admin
```

#### Step 4: Verify Build

Check that binaries are created:
```bash
# Windows
dir bin\

# Linux/Mac
ls -la bin/
```

Expected output:
```
server.exe  (or server on Linux/Mac)
client.exe  (or client on Linux/Mac)
admin.exe   (or admin on Linux/Mac)
```

### 4.3 Network Configuration

**Firewall Settings:**
- Ensure port 8080 is open (or change port in code)
- For Windows Firewall:
  ```powershell
  New-NetFirewallRule -DisplayName "RPC Server" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
  ```

**Multi-Computer Setup:**
1. **Server Machine**: Run server on IP address (e.g., 192.168.1.100)
2. **Client Machines**: Connect using server IP:
   ```bash
   ./client -server 192.168.1.100:8080 -id client1
   ```

### 4.4 Running the System

#### Start Server

**Terminal 1:**
```bash
# Windows
.\bin\server.exe

# Linux/Mac
./bin/server
```

Expected output:
```
Remote Shell RPC Server started on :8080
Waiting for clients...
```

#### Start Clients

**Terminal 2 (Client 1):**
```bash
# Windows
.\bin\client.exe -id client1

# Linux/Mac
./bin/client -id client1
```

**Terminal 3 (Client 2):**
```bash
.\bin\client.exe -id client2
```

#### Check Active Clients

**Terminal 4 (Admin):**
```bash
.\bin\admin.exe
```

Expected output:
```
Active clients (2):
  1. client1
  2. client2
```

### 4.5 Troubleshooting

**Problem: "go is not recognized"**
- Solution: Install Go and restart terminal
- Verify: `go version`

**Problem: "port 8080 already in use"**
- Solution: Change port in `server/main.go` line 203
- Or kill process using port 8080

**Problem: "connection refused"**
- Solution: Ensure server is running
- Check firewall settings
- Verify server IP address

**Problem: "Access denied" for PowerShell scripts**
- Solution:
  ```powershell
  Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
  ```

---

## 5. Usage Guideline

### 5.1 Basic Usage

#### Starting Server
```bash
.\bin\server.exe
```

#### Connecting as Client
```bash
.\bin\client.exe -id my-client
```

#### Running Single Command (Non-Interactive)
```bash
.\bin\client.exe -id my-client -cmd "dir"
.\bin\client.exe -id my-client -cmd "echo Hello World"
```

### 5.2 Interactive Mode Commands

Once connected, you can use these commands:

#### Shell Commands
```
[client1@remote]$ dir              # List files (Windows)
[client1@remote]$ echo Hello       # Print text
[client1@remote]$ pwd              # Print working directory
```

#### Special Commands
```
[client1@remote]$ help             # Show help
[client1@remote]$ cd C:\Windows    # Change directory
[client1@remote]$ setenv MY_VAR hello  # Set environment variable
[client1@remote]$ exit             # Disconnect
```

### 5.3 Usage Examples

#### Example 1: Basic Command Execution
```
[client1@remote]$ dir
 Volume in drive C has no label.
 Volume Serial Number is XXXX-XXXX

 Directory of C:\Users\Admin\Downloads\abcd

[client1@remote]$ echo Hello World
Hello World
```

#### Example 2: Change Directory
```
[client1@remote]$ cd C:\Windows
Directory changed
[client1@remote]$ dir
...
```

#### Example 3: Environment Variables
```
[client1@remote]$ setenv MY_NAME John
Set MY_NAME=John
[client1@remote]$ echo %MY_NAME%
John
```

#### Example 4: Multiple Clients
**Terminal 1 (Server):**
```
Remote Shell RPC Server started on :8080
Waiting for clients...
New client connected: 127.0.0.1:xxxxx
[Client client1] Registered (new session)
[Client client1] Executed: dir (Exit: 0)
New client connected: 127.0.0.1:xxxxx
[Client client2] Registered (new session)
[Client client2] Executed: echo Hello (Exit: 0)
```

**Terminal 2 (Client 1):**
```
[client1@remote]$ dir
...
```

**Terminal 3 (Client 2):**
```
[client2@remote]$ echo Hello
Hello
```

**Terminal 4 (Admin):**
```
Active clients (2):
  1. client1
  2. client2
```

### 5.4 Advanced Usage

#### Connecting to Remote Server
```bash
.\bin\client.exe -server 192.168.1.100:8080 -id client1
```

#### Auto-Generate Client ID
```bash
.\bin\client.exe
# Will generate: client-1234567890123456789
```

#### Scripting Example
```bash
# Execute multiple commands
.\bin\client.exe -id script-client -cmd "dir"
.\bin\client.exe -id script-client -cmd "echo Done"
```

---

## 6. Results Captured as Images

### 6.1 Screenshot Requirements

Bạn cần chụp screenshots cho các phần sau:

#### 6.1.1 Server Startup
- Screenshot: Server đang chạy và chờ clients
- Location: Terminal showing server logs

#### 6.1.2 Client Connection
- Screenshot: Client kết nối thành công
- Location: Client terminal showing "Connected to server..."

#### 6.1.3 Command Execution
- Screenshot: Thực thi lệnh `dir` hoặc `ls`
- Location: Client terminal showing command output

#### 6.1.4 Multiple Clients
- Screenshot: 2-3 clients cùng kết nối
- Location: Multiple terminals

#### 6.1.5 Admin Tool - List Clients
- Screenshot: Admin tool hiển thị danh sách clients
- Location: Admin terminal

#### 6.1.6 Change Directory
- Screenshot: Client thay đổi directory và list files
- Location: Client terminal

#### 6.1.7 Environment Variables
- Screenshot: Set và sử dụng environment variable
- Location: Client terminal

#### 6.1.8 Server Logs
- Screenshot: Server logs hiển thị multiple clients
- Location: Server terminal

### 6.2 How to Take Screenshots

**Windows:**
- `Win + Shift + S`: Snipping Tool
- `Alt + Print Screen`: Screenshot active window

**Linux:**
- `Print Screen`: Full screen
- `Shift + Print Screen`: Select area

**macOS:**
- `Cmd + Shift + 4`: Select area
- `Cmd + Shift + 3`: Full screen

### 6.3 Image Organization

Tạo folder `screenshots/` và đặt tên files:
```
screenshots/
├── 01-server-startup.png
├── 02-client-connection.png
├── 03-command-execution.png
├── 04-multiple-clients.png
├── 05-admin-list-clients.png
├── 06-change-directory.png
├── 07-environment-variables.png
└── 08-server-logs.png
```

**Note**: Chèn các screenshots này vào report khi submit.

---

## 7. Contribution of Each Team Member

### 7.1 Contribution Template

**Member 1: [Name]**
- **Role**: [e.g., Lead Developer, Server Implementation]
- **Contributions**:
  - Implemented RPC server with concurrent client handling (30%)
  - Designed session management system (15%)
  - Created deployment documentation (10%)
- **Total**: 55%

**Member 2: [Name]**
- **Role**: [e.g., Client Developer, Testing]
- **Contributions**:
  - Implemented RPC client with interactive mode (25%)
  - Created admin tool (10%)
  - Testing and bug fixes (10%)
- **Total**: 45%

**Member 3: [Name]**
- **Role**: [e.g., Documentation, Protocol Design]
- **Contributions**:
  - Protocol design and documentation (20%)
  - Created usage guidelines (15%)
  - Presentation slides (10%)
- **Total**: 45%

**Note**: Tổng contribution phải = 100% (hoặc điều chỉnh theo số lượng members)

### 7.2 How to Calculate

- **Code Lines**: Số dòng code mỗi người viết
- **Features**: Features mỗi người implement
- **Documentation**: Phần documentation mỗi người viết
- **Testing**: Test cases và bug fixes

---

## 8. Conclusion

### 8.1 Achievements
- ✅ Successfully implemented RPC-based remote shell system
- ✅ Support for multiple concurrent clients
- ✅ Session management with working directory and environment variables
- ✅ Cross-platform compatibility (Windows, Linux, macOS)
- ✅ Efficient concurrent processing with goroutines

### 8.2 Future Improvements
- [ ] Real-time streaming output (instead of batch)
- [ ] Interactive TTY support
- [ ] Authentication and authorization
- [ ] TLS/SSL encryption
- [ ] Command history
- [ ] File transfer (scp-like)
- [ ] gRPC instead of net/rpc
- [ ] Metrics and monitoring dashboard

### 8.3 Lessons Learned
- Understanding RPC protocols in practice
- Concurrent programming with goroutines
- Session management in distributed systems
- Network programming with TCP/IP
- Error handling in distributed systems

---

## Appendix

### A. Project Structure
```
remote-shell-rpc/
├── server/
│   └── main.go          # RPC Server implementation
├── client/
│   └── main.go          # RPC Client implementation
├── admin/
│   └── main.go          # Admin tool
├── go.mod               # Go module definition
├── README.md            # Project documentation
├── SETUP.md             # Setup instructions
├── build.ps1            # PowerShell build script
├── build.bat            # Batch build script
├── run-server.ps1       # Server runner script
├── run-client.ps1       # Client runner script
└── MIDTERM_REPORT.md    # This report
```

### B. Code Statistics
- **Total Lines of Code**: ~600 lines
- **Server Code**: ~240 lines
- **Client Code**: ~200 lines
- **Admin Code**: ~30 lines
- **Documentation**: ~1000+ lines

### C. References
- Go RPC Documentation: https://pkg.go.dev/net/rpc
- Go Concurrency: https://go.dev/tour/concurrency/1
- Kubernetes kubectl exec: https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#exec

---

**End of Report**

