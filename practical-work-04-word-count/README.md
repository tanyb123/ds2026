# Practical Work 4: Word Count (MapReduce)

## Mô tả
Implement Word Count example sử dụng MapReduce.

## Implementation
Tự implement MapReduce framework vì không có framework phổ biến cho Python (Hadoop là Java).

## Sử dụng
```bash
# Tạo input files
mkdir -p input output
echo "hello world hello" > input/file1.txt
echo "world hello" > input/file2.txt

# Run word count
python3 wordcount.py input/ output/ --mappers 2

# Xem kết quả
cat output/part-00000
```

## MapReduce Design
- **Map**: Tokenize và emit `<word, 1>` cho mỗi word
- **Reduce**: Sum tất cả values cho mỗi word
- **Parallel**: Sử dụng multiprocessing Pool
