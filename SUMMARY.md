# Tổng kết Practical Works

## ✅ Đã hoàn thành 4 bài practice bằng Python

### Bài 1: TCP File Transfer
- **Files**: `server.py`, `client.py`
- **Tech**: Python, TCP sockets
- **Chức năng**: Chuyển file qua TCP/IP

### Bài 2: RPC File Transfer  
- **Files**: `server.py`, `client.py`, `filetransfer.proto`
- **Tech**: Python, gRPC
- **Chức năng**: Chuyển file qua RPC với streaming

### Bài 3: MPI File Transfer
- **Files**: `mpi_file_transfer.py`
- **Tech**: Python, mpi4py
- **Chức năng**: Chuyển file song song với MPI

### Bài 4: Word Count (MapReduce)
- **Files**: `wordcount.py`
- **Tech**: Python, multiprocessing
- **Chức năng**: Đếm từ sử dụng MapReduce

## Cách push lên GitHub

1. **Config Git** (chỉ cần làm 1 lần):
   ```bash
   git config --global user.name "Nguyễn Duy Tân"
   git config --global user.email "your.email@example.com"
   ```

2. **Add remote**:
   ```bash
   git remote add origin https://github.com/tanyb123/ds2026.git
   ```

3. **Add và commit**:
   ```bash
   git add practical-work-*/ README.md requirements.txt .gitignore
   git commit -m "Add practical works 1-4: TCP, RPC, MPI file transfer and Word Count MapReduce (Python)"
   ```

4. **Push**:
   ```bash
   git push -u origin main
   ```

Hoặc dùng script:
- Windows: `.\push_to_github.ps1`
- Linux/Mac: `./push_to_github.sh`

## Cài đặt dependencies

```bash
pip install -r requirements.txt
```

## Test các bài

Xem README.md trong từng thư mục `practical-work-XX-*/` để biết cách chạy.

