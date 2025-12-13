# Practical Work 3: MPI File Transfer

## Mô tả
Nâng cấp TCP File Transfer sang MPI.

## Cài đặt
```bash
pip install mpi4py
# Cần cài OpenMPI hoặc MPICH
# Ubuntu: sudo apt-get install openmpi-bin libopenmpi-dev
```

## Sử dụng
```bash
# Run với 4 processes
mpirun -np 4 python3 mpi_file_transfer.py --send-file test.txt
```

## Implementation
Chọn **OpenMPI** vì phổ biến, cross-platform, và dễ sử dụng với Python (mpi4py).
