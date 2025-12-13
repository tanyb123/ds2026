# Hướng dẫn Cài đặt và Sử dụng

## Bước 1: Cài đặt Go

Nếu bạn chưa có Go, hãy cài đặt:

1. Tải Go từ: https://go.dev/dl/
2. Chọn bản Windows (ví dụ: `go1.21.x.windows-amd64.msi`)
3. Chạy installer và làm theo hướng dẫn
4. **Quan trọng**: Sau khi cài đặt, **khởi động lại PowerShell** để PATH được cập nhật

### Kiểm tra cài đặt

Mở PowerShell mới và chạy:
```powershell
go version
```

Nếu thấy version (ví dụ: `go version go1.21.5 windows/amd64`) thì đã cài đặt thành công!

## Bước 2: Build Project

### Cách 1: Dùng PowerShell script (Khuyên dùng)
```powershell
.\build.ps1
```

### Cách 2: Dùng batch file
```powershell
.\build.bat
```

### Cách 3: Build thủ công
```powershell
go build -o bin\server.exe ./server
go build -o bin\client.exe ./client
go build -o bin\admin.exe ./admin
```

## Bước 3: Chạy Hệ thống

### Terminal 1: Chạy Server
```powershell
# Cách 1: Dùng script
.\run-server.ps1

# Cách 2: Chạy trực tiếp
.\bin\server.exe
```

### Terminal 2: Chạy Client 1
```powershell
# Cách 1: Dùng script
.\run-client.ps1 -ClientID client1

# Cách 2: Chạy trực tiếp
.\bin\client.exe -id client1

# Hoặc để tự động tạo ID
.\bin\client.exe
```

### Terminal 3: Chạy Client 2 (để test multiple clients)
```powershell
.\bin\client.exe -id client2
```

### Terminal 4: Xem danh sách Clients
```powershell
.\bin\admin.exe
```

## Ví dụ Sử dụng

### Trong Client Terminal:
```
[client1@remote]$ pwd
C:\Users\Admin\Downloads\abcd

[client1@remote]$ dir
...

[client1@remote]$ cd C:\Windows
Directory changed

[client1@remote]$ setenv MY_VAR hello
Set MY_VAR=hello

[client1@remote]$ echo %MY_VAR%
hello

[client1@remote]$ exit
```

## Troubleshooting

### Lỗi: "go is not recognized"
- **Nguyên nhân**: Go chưa được cài đặt hoặc chưa có trong PATH
- **Giải pháp**: 
  1. Cài đặt Go từ https://go.dev/dl/
  2. Khởi động lại PowerShell
  3. Kiểm tra lại: `go version`

### Lỗi: "cannot find package"
- **Nguyên nhân**: Đang ở sai thư mục
- **Giải pháp**: Đảm bảo bạn đang ở thư mục `abcd` (root của project)

### Lỗi: "Access denied" khi chạy script
- **Nguyên nhân**: PowerShell execution policy
- **Giải pháp**: Chạy lệnh này một lần:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

## Cấu trúc Project

```
abcd/
├── server/
│   └── main.go          # RPC Server
├── client/
│   └── main.go          # RPC Client
├── admin/
│   └── main.go          # Admin tool
├── go.mod               # Go module
├── build.ps1            # PowerShell build script
├── build.bat            # Batch build script
├── run-server.ps1       # Run server script
├── run-client.ps1       # Run client script
└── README.md            # Documentation
```




