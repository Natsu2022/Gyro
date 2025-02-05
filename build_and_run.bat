@echo off
go build -o main.exe main.go
if %errorlevel% neq 0 exit /b %errorlevel%
main.exe