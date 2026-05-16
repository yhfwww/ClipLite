@echo off
echo Building ClipLite for Windows...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

go build -ldflags="-H windowsgui -s -w" -o ClipLite.exe .

echo Build complete: ClipLite.exe
pause
