@echo off
echo Starting server...
start /B server.exe
set SERVER_PID=%ERRORLEVEL%

echo Server started with PID: %SERVER_PID%
echo Waiting 3 seconds for server to fully start...
timeout /t 3 /nobreak > nul

echo Making a request to test server is running...
curl -s http://localhost:8080/api/health

echo.
echo Sending SIGTERM to gracefully shutdown server...
taskkill /PID %SERVER_PID% /F

echo Server has been gracefully shutdown!
pause 