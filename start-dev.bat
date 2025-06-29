@echo off
echo Starting No Heroes, No Lies Development Environment...
echo.

echo Starting Backend Server...
start "Backend" cmd /k "cd backend && set ALLOWED_DOMAIN_SUFFIX=localhost:8080 && go run ./cmd/server/main.go"

echo Waiting for backend to start...
timeout /t 3 /nobreak > nul

echo Starting Frontend Development Server...
start "Frontend" cmd /k "cd frontend && npm run dev"

echo.
echo Development servers are starting...
echo Backend: http://localhost:8080
echo Frontend: http://localhost:3000
echo.
echo Press any key to exit this script (servers will continue running)
pause > nul 