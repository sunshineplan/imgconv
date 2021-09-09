@echo off

cd converter
go build -ldflags "-s -w" -o ../converter.exe
cd ..
