@echo off
set ALLOWED_DOMAIN_SUFFIX=localhost:8080
cd backend
go run ./cmd/server/main.go 