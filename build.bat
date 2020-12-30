@echo off

go build -ldflags "-s -w" -o converter.exe ./cmd 
