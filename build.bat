@echo off
cd cli
go build -ldflags "-s -w" -o ../convert.exe
cd ..
