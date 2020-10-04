@echo off
cd cmd
go build -ldflags "-s -w" -o ../converter.exe
cd ..
