@echo off
echo === Checking Go Configuration ===
echo.
echo Go Version:
go version
echo.
echo GOROOT (should be C:\Program Files\Go):
go env GOROOT
echo.
echo Testing compile...
cd chain\x\auth\keeper
go test -v -run TestGetParams -timeout 30s
