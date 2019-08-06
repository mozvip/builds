@ECHO OFF
rsrc -manifest awi.exe.manifest -o awi.syso
go build -o awi.exe