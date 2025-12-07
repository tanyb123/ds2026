# Hướng dẫn Push lên GitHub

## Bước 1: Cấu hình Git (BẮT BUỘC)

```bash
# Windows PowerShell
git config --global user.name "Nguyễn Duy Tân"
git config --global user.email "your.email@example.com"

# Hoặc Linux/Mac
git config --global user.name "Nguyễn Duy Tân"
git config --global user.email "your.email@example.com"
```

## Bước 2: Add remote repository

```bash
git remote add origin https://github.com/tanyb123/ds2026.git
```

## Bước 3: Add và commit files

```bash
# Add tất cả files
git add .

# Commit với message
git commit -m "Add practical works 1-4: TCP, RPC, MPI file transfer and Word Count MapReduce"

# Hoặc commit từng bài
git add practical-work-01-tcp-file-transfer/
git commit -m "Practical Work 1: TCP File Transfer"

git add practical-work-02-rpc-file-transfer/
git commit -m "Practical Work 2: RPC File Transfer"

git add practical-work-03-mpi-file-transfer/
git commit -m "Practical Work 3: MPI File Transfer"

git add practical-work-04-word-count/
git commit -m "Practical Work 4: Word Count MapReduce"
```

## Bước 4: Push lên GitHub

### Cách 1: Dùng script (Khuyến nghị)

**Windows PowerShell:**
```powershell
.\push_to_github.ps1
```

**Linux/Mac:**
```bash
chmod +x push_to_github.sh
./push_to_github.sh
```

### Cách 2: Manual

```bash
# Push lần đầu
git push -u origin main

# Hoặc nếu branch là master
git push -u origin master

# Các lần sau chỉ cần
git push
```

## Quick Commands (Copy & Paste)

```bash
# 1. Config git (chỉ cần làm 1 lần)
git config --global user.name "Nguyễn Duy Tân"
git config --global user.email "your.email@example.com"

# 2. Add remote
git remote add origin https://github.com/tanyb123/ds2026.git

# 3. Add files
git add practical-work-*/ README.md requirements.txt .gitignore

# 4. Commit
git commit -m "Add practical works 1-4: TCP, RPC, MPI file transfer and Word Count MapReduce (Python)"

# 5. Push
git push -u origin main
```

## Lưu ý

- Đảm bảo đã có quyền truy cập vào repo https://github.com/tanyb123/ds2026
- Nếu repo yêu cầu authentication, có thể cần dùng Personal Access Token

