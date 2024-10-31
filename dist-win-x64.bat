mkdir dist
rmdir /S /Q dist\win-x64
mkdir dist\win-x64

mkdir dist\win-x64\command
copy command\command.exe dist\win-x64\command\BCSPanel.exe

mkdir dist\win-x64\frontend\dist
copy frontend\dist dist\win-x64\frontend\dist

copy BCSPanel.exe dist\win-x64\BCSPanelServer.exe

pause
