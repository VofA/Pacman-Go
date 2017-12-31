@echo off

echo Delete directory: Build
RD /S /Q Build

echo Create directory: Build
md Build

echo | set /p=Copy resources: 
xcopy Source\Data Build\Data\ /Q /E

echo | set /p=Build source code: 
go build -o Build/Build.exe Source/main.go
echo Ok

pause