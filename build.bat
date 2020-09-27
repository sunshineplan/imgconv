@echo off
cd cmd
go build -ldflags "-s -w" -o ../convert.exe
cd ..
