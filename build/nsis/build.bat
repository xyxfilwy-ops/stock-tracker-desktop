@echo off
setlocal

cd /d "E:\kimi项目\股票基金跟踪\stock-tracker-desktop\build\nsis"
"C:\Program Files (x86)\NSIS\makensis.exe" StockTracker.nsi > nsis_build.log 2>&1

echo Return code: %ERRORLEVEL%

if exist "StockTracker-1.0.0-windows-amd64-installer.exe" (
    echo INSTALLER_CREATED
    dir "StockTracker-1.0.0-windows-amd64-installer.exe"
) else (
    echo INSTALLER_NOT_CREATED
)

pause
