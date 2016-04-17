@echo off
set onecpath=c:\Program Files (x86)\1cv8\common\
set ib=%gopath%\src\github.com\mcarrowd\oneclogbeat\testing\infobase\
set processors=%gopath%\src\github.com\mcarrowd\oneclogbeat\testing\dataprocessors\
set template=%gopath%\src\github.com\mcarrowd\oneclogbeat\testing\1Cv8.dt
del /Q /S %ib%
"%onecpath%\1cestart.exe" CREATEINFOBASE File="%ib%" /UseTemplate "%template%"
"%onecpath%\1cestart.exe" ENTERPRISE /WA- /N "Администратор" /F"%ib%" /DisableStartupMessage /Execute "%processors%\WriteEventLog.epf"