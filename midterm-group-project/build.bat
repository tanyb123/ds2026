@echo off
echo Building Remote Shell RPC System...

if not exist bin mkdir bin

echo Building server...
go build -o bin\server.exe ./server
if errorlevel 1 (
    echo Failed to build server
    exit /b 1
)

echo Building client...
go build -o bin\client.exe ./client
if errorlevel 1 (
    echo Failed to build client
    exit /b 1
)

echo Building admin tool...
go build -o bin\admin.exe ./admin
if errorlevel 1 (
    echo Failed to build admin
    exit /b 1
)

echo.
echo Build completed successfully!
echo Binaries are in the bin\ directory
echo.
echo To run:
echo   Server:  bin\server.exe
echo   Client:  bin\client.exe -id my-client
echo   Admin:   bin\admin.exe




