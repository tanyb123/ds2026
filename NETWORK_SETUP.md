# Network Setup Guide - Multi-Computer Connection

Hướng dẫn cấu hình để kết nối nhiều máy tính qua internet (không cần VPN).

---

## Câu hỏi: Có thể kết nối qua internet không?

**Trả lời: CÓ!** Hệ thống hỗ trợ kết nối qua internet, không cần VPN nếu cấu hình đúng.

---

## Các Tình Huống Kết Nối

### Tình huống 1: Tất cả cùng mạng LAN (Local Network)
```
Computer A (Server)          Computer B (Client)        Computer C (Client)
192.168.1.100               192.168.1.101             192.168.1.102
     │                            │                          │
     └────────────────────────────┴──────────────────────────┘
                          Router (192.168.1.1)
```

**Cách kết nối:**
- Server: `.\bin\server.exe` (bind to 0.0.0.0:8080)
- Client B: `.\bin\client.exe -server 192.168.1.100:8080 -id client1`
- Client C: `.\bin\client.exe -server 192.168.1.100:8080 -id client2`

**Ưu điểm:**
- ✅ Không cần cấu hình đặc biệt
- ✅ Kết nối nhanh, độ trễ thấp
- ✅ Không cần public IP

---

### Tình huống 2: Khác mạng LAN - Qua Internet (Public IP)
```
Computer A (Server)          Computer B (Client)        Computer C (Client)
Public IP: 203.0.113.1      Public IP: 198.51.100.1   Public IP: 192.0.2.1
     │                            │                          │
     └────────────────────────────┴──────────────────────────┘
                          Internet
```

**Cách kết nối:**
- Server: `.\bin\server.exe` (bind to 0.0.0.0:8080)
- Client B: `.\bin\client.exe -server 203.0.113.1:8080 -id client1`
- Client C: `.\bin\client.exe -server 203.0.113.1:8080 -id client2`

**Yêu cầu:**
- ✅ Server phải có public IP
- ✅ Firewall mở port 8080
- ✅ Router port forwarding (nếu server sau NAT)

**Ưu điểm:**
- ✅ Không cần VPN
- ✅ Kết nối từ bất kỳ đâu
- ✅ Không cần cùng mạng

---

### Tình huống 3: Server sau NAT/Router (Private IP)
```
Computer A (Server)          Computer B (Client)        Computer C (Client)
Private: 192.168.1.100      Public IP: 198.51.100.1   Public IP: 192.0.2.1
Public: 203.0.113.1 (Router)
     │                            │                          │
     └────────────────────────────┴──────────────────────────┘
                          Internet
```

**Cách kết nối:**
- Server: `.\bin\server.exe` (bind to 0.0.0.0:8080)
- Router: Port forwarding 8080 → 192.168.1.100:8080
- Client B: `.\bin\client.exe -server 203.0.113.1:8080 -id client1`
- Client C: `.\bin\client.exe -server 203.0.113.1:8080 -id client2`

**Yêu cầu:**
- ✅ Router có public IP
- ✅ Port forwarding cấu hình đúng
- ✅ Firewall mở port 8080

---

### Tình huống 4: Tất cả sau NAT (Cần VPN hoặc Tunneling)
```
Computer A (Server)          Computer B (Client)        Computer C (Client)
Private: 192.168.1.100      Private: 192.168.2.100    Private: 192.168.3.100
     │                            │                          │
     └────────────────────────────┴──────────────────────────┘
                    VPN Server / Tunneling Service
```

**Giải pháp:**
1. **VPN**: Tất cả kết nối VPN, sau đó dùng private IP
2. **ngrok/Cloudflare Tunnel**: Tạo tunnel cho server
3. **SSH Tunnel**: Tạo SSH tunnel

**Khi nào cần VPN:**
- ❌ Không có public IP
- ❌ Không thể port forward
- ❌ Cần bảo mật cao hơn
- ❌ Firewall chặn port 8080

---

## Hướng Dẫn Cấu Hình Chi Tiết

### Bước 1: Kiểm tra IP của Server

**Windows:**
```powershell
ipconfig
# Tìm IPv4 Address (ví dụ: 192.168.1.100)
```

**Linux/Mac:**
```bash
ifconfig
# hoặc
ip addr show
```

### Bước 2: Cấu hình Server

Server hiện tại bind to `0.0.0.0:8080` (tất cả interfaces), nên đã sẵn sàng nhận kết nối từ bất kỳ đâu.

**Kiểm tra server đang listen:**
```powershell
# Windows
netstat -an | findstr 8080

# Linux/Mac
netstat -an | grep 8080
# hoặc
ss -tlnp | grep 8080
```

### Bước 3: Cấu hình Firewall

#### Windows Firewall
```powershell
# Mở port 8080
New-NetFirewallRule -DisplayName "RPC Server" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
```

Hoặc qua GUI:
1. Windows Defender Firewall → Advanced Settings
2. Inbound Rules → New Rule
3. Port → TCP → 8080 → Allow

#### Linux Firewall (ufw)
```bash
sudo ufw allow 8080/tcp
sudo ufw reload
```

#### Linux Firewall (iptables)
```bash
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
```

### Bước 4: Port Forwarding (Nếu server sau NAT)

**Router Configuration:**
1. Đăng nhập router (thường 192.168.1.1)
2. Tìm "Port Forwarding" hoặc "Virtual Server"
3. Thêm rule:
   - External Port: 8080
   - Internal IP: 192.168.1.100 (IP của server)
   - Internal Port: 8080
   - Protocol: TCP

**Lưu ý:**
- External IP của router = Public IP (kiểm tra tại whatismyip.com)
- Clients sẽ kết nối đến Public IP này

### Bước 5: Kết nối từ Clients

**Từ máy khác mạng:**
```bash
# Thay SERVER_PUBLIC_IP bằng IP thực tế
.\bin\client.exe -server SERVER_PUBLIC_IP:8080 -id client1
```

**Ví dụ:**
```bash
# Server có public IP: 203.0.113.1
.\bin\client.exe -server 203.0.113.1:8080 -id client1

# Hoặc server sau NAT, router có IP: 203.0.113.1
.\bin\client.exe -server 203.0.113.1:8080 -id client1
```

---

## Cải Tiến Code để Hỗ Trợ Tốt Hơn

### Option 1: Thêm flag để bind specific IP

Có thể cải thiện server để bind specific IP:

```go
// server/main.go
var bindAddr = flag.String("bind", "0.0.0.0:8080", "Address to bind server")

func main() {
    flag.Parse()
    listener, err := net.Listen("tcp", *bindAddr)
    // ...
}
```

### Option 2: Hiển thị IP khi server start

Thêm code để hiển thị IP addresses khi server start:

```go
// Hiển thị tất cả IP addresses
addrs, _ := net.InterfaceAddrs()
for _, addr := range addrs {
    log.Printf("Server listening on: %s:8080", addr)
}
```

---

## Kiểm Tra Kết Nối

### Test từ Client Machine

**Kiểm tra có thể kết nối đến server:**
```bash
# Windows
Test-NetConnection -ComputerName SERVER_IP -Port 8080

# Linux/Mac
telnet SERVER_IP 8080
# hoặc
nc -zv SERVER_IP 8080
```

### Test từ Server

**Kiểm tra server đang listen:**
```bash
# Windows
netstat -an | findstr 8080

# Linux/Mac
ss -tlnp | grep 8080
```

---

## Giải Pháp Thay Thế (Nếu Không Có Public IP)

### 1. Sử dụng ngrok (Dễ nhất)

**Trên Server:**
```bash
# Download ngrok: https://ngrok.com/
ngrok tcp 8080
```

**Output:**
```
Forwarding   tcp://0.tcp.ngrok.io:12345 -> localhost:8080
```

**Trên Client:**
```bash
.\bin\client.exe -server 0.tcp.ngrok.io:12345 -id client1
```

**Ưu điểm:**
- ✅ Không cần public IP
- ✅ Không cần port forwarding
- ✅ Dễ setup
- ❌ Free plan có giới hạn

### 2. Sử dụng Cloudflare Tunnel

Tương tự ngrok, nhưng miễn phí và không giới hạn.

### 3. Sử dụng VPN

**Setup VPN Server:**
- OpenVPN
- WireGuard
- Tailscale (dễ nhất)

Sau khi tất cả kết nối VPN, dùng private IP như mạng LAN.

---

## Tóm Tắt

| Tình huống | Cần VPN? | Cần Public IP? | Cần Port Forward? |
|------------|----------|----------------|-------------------|
| Cùng LAN | ❌ Không | ❌ Không | ❌ Không |
| Server có Public IP | ❌ Không | ✅ Có | ❌ Không |
| Server sau NAT | ❌ Không | ✅ Router có | ✅ Có |
| Tất cả sau NAT | ✅ Có | ❌ Không | ❌ Không |

**Kết luận:**
- ✅ **Có thể kết nối qua internet** nếu server có public IP hoặc port forwarding
- ❌ **Chỉ cần VPN** nếu không có public IP và không thể port forward
- ✅ **Hệ thống hiện tại đã hỗ trợ** kết nối qua internet

---

## Ví Dụ Thực Tế

### Scenario: 3 máy khác mạng

**Máy 1 (Server):**
- IP: 203.0.113.1 (public IP)
- Chạy: `.\bin\server.exe`
- Firewall: Mở port 8080

**Máy 2 (Client 1):**
- IP: 198.51.100.1 (khác mạng)
- Chạy: `.\bin\client.exe -server 203.0.113.1:8080 -id client1`

**Máy 3 (Client 2):**
- IP: 192.0.2.1 (khác mạng)
- Chạy: `.\bin\client.exe -server 203.0.113.1:8080 -id client2`

**Kết quả:** ✅ Cả 3 máy kết nối được, không cần VPN!

---

## Troubleshooting

**Lỗi: "connection refused"**
- Kiểm tra server đang chạy
- Kiểm tra firewall
- Kiểm tra IP address đúng

**Lỗi: "timeout"**
- Kiểm tra port forwarding
- Kiểm tra firewall trên router
- Kiểm tra server có public IP không

**Lỗi: "no route to host"**
- Kiểm tra network connectivity
- Ping server IP
- Kiểm tra routing

---

**Kết luận: Hệ thống đã hỗ trợ kết nối qua internet, không cần VPN nếu cấu hình đúng!**



