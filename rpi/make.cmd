@echo off
setlocal
  mkdir build

  set GOOS=linux
  set GOARCH=arm
  set GOARM=5

  go build -o build\rpi
endlocal