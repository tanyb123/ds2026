# Kết Quả Kiểm Tra IP

## Thông Tin Mạng Của Bạn

### Local IP (Private IP)
- **IPv4 Address**: `172.20.10.8`
- **Subnet Mask**: `255.255.255.240`
- **Default Gateway**: `172.20.10.1`

### Public IP (Từ Internet)
- **Public IP**: `171.255.114.251`

## Phân Tích

### ❌ Máy bạn KHÔNG có public IP trực tiếp

**Lý do:**
- Local IP (`172.20.10.8`) ≠ Public IP (`171.255.114.251`)
- IP `172.20.10.8` là **private IP** (thuộc dải 172.16.0.0/12)
- Máy bạn đang **sau NAT/Router**

### Kết Luận

Máy bạn đang kết nối qua:
- **Router/Gateway**: `172.20.10.1`
- **Public IP của Router**: `171.255.114.251`
- **Mạng**: Có thể là mobile hotspot hoặc router WiFi

---

## Cách Kết Nối Từ Máy Khác

### Option 1: Port Forwarding (Khuyên dùng)

**Bước 1: Cấu hình Port Forwarding trên Router**
1. Đăng nhập router (thường tại `172.20.10.1`)
2. Tìm "Port Forwarding" hoặc "Virtual Server"
3. Thêm rule:
   - **External Port**: 8080
   - **Internal IP**: 172.20.10.8
   - **Internal Port**: 8080
   - **Protocol**: TCP

**Bước 2: Clients kết nối**
```powershell
# Clients từ máy khác kết nối đến:
.\bin\client.exe -server 171.255.114.251:8080 -id client1
```

**Lưu ý:**
- Nếu là mobile hotspot, có thể không hỗ trợ port forwarding
- Cần kiểm tra router có hỗ trợ không

---

### Option 2: Sử dụng ngrok (Dễ nhất - Không cần port forwarding)

**Trên máy Server (máy bạn):**

1. Download ngrok: https://ngrok.com/download
2. Chạy:
```powershell
ngrok tcp 8080
```

**Output sẽ như:**
```
Forwarding   tcp://0.tcp.ngrok.io:12345 -> localhost:8080
```

**Trên máy Client:**
```powershell
.\bin\client.exe -server 0.tcp.ngrok.io:12345 -id client1
```

**Ưu điểm:**
- ✅ Không cần port forwarding
- ✅ Không cần cấu hình router
- ✅ Hoạt động với mọi loại mạng
- ❌ Free plan có giới hạn

---

### Option 3: Sử dụng VPN

Nếu tất cả máy đều kết nối VPN, có thể dùng private IP:
- Server: `172.20.10.8:8080`
- Clients: Kết nối VPN trước, sau đó dùng IP VPN

---

## Khuyến Nghị

### Cho Midterm Project:

**Nếu test trên cùng mạng LAN:**
- ✅ Dùng local IP: `172.20.10.8:8080`
- ✅ Không cần port forwarding
- ✅ Dễ nhất

**Nếu test từ mạng khác:**
- ✅ **Dùng ngrok** (dễ nhất, không cần cấu hình)
- ⚠️ Port forwarding (nếu router hỗ trợ)
- ⚠️ VPN (nếu có)

---

## Hướng Dẫn Sử Dụng ngrok

### Bước 1: Download và Setup
1. Vào https://ngrok.com/
2. Đăng ký tài khoản miễn phí
3. Download ngrok cho Windows
4. Giải nén vào thư mục dự án

### Bước 2: Chạy ngrok
```powershell
# Terminal 1: Chạy server
.\bin\server.exe

# Terminal 2: Chạy ngrok
.\ngrok.exe tcp 8080
```

### Bước 3: Lấy URL
ngrok sẽ hiển thị:
```
Forwarding   tcp://0.tcp.ngrok.io:12345 -> localhost:8080
```

### Bước 4: Clients kết nối
```powershell
# Từ máy khác (hoặc cùng máy)
.\bin\client.exe -server 0.tcp.ngrok.io:12345 -id client1
```

---

## Tóm Tắt

| Tình Huống | Cách Kết Nối | Khó Khăn |
|------------|--------------|----------|
| Cùng LAN | `172.20.10.8:8080` | ✅ Dễ |
| Khác mạng + Port Forward | `171.255.114.251:8080` | ⚠️ Cần cấu hình router |
| Khác mạng + ngrok | `0.tcp.ngrok.io:xxxxx` | ✅ Dễ nhất |
| VPN | IP VPN | ⚠️ Cần VPN |

**Kết luận cho bạn:**
- ❌ Không có public IP trực tiếp
- ✅ Có thể dùng ngrok (khuyên dùng)
- ✅ Hoặc port forwarding nếu router hỗ trợ
- ✅ Hoặc test trên cùng LAN

---

**Bạn muốn tôi hướng dẫn setup ngrok không?** 🚀



