@ECHO OFF
rsrc -manifest builds.exe.manifest -o builds.syso
go build -o builds.exe