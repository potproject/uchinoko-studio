@echo off
cd front && pnpm install && pnpm build
cd ../server && go run main.go
