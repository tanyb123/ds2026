# PowerShell script to push to GitHub

# Add all practical work files
git add practical-work-01-tcp-file-transfer/
git add practical-work-02-rpc-file-transfer/
git add practical-work-03-mpi-file-transfer/
git add practical-work-04-word-count/
git add README.md
git add requirements.txt
git add .gitignore
git add GIT_SETUP.md

# Commit
git commit -m "Add practical works 1-4: TCP, RPC, MPI file transfer and Word Count MapReduce (Python)"

# Add remote if not exists
git remote add origin https://github.com/tanyb123/ds2026.git 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Remote already exists or error occurred"
}

# Push
git push -u origin main
if ($LASTEXITCODE -ne 0) {
    git push -u origin master
}

