@echo off
setlocal
  set GOOS=linux
  set GOARCH=arm
  set GOARM=5

  go build
endlocal