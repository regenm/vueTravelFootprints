@echo off
cd /d "%~dp0"
if not exist data mkdir data
if not exist uploads mkdir uploads
echo 旅迹服务启动中（本机 127.0.0.1:5000，对外请走 Caddy）...
travel-server.exe
pause
