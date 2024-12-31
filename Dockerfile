FROM node:20-alpine AS frontend-builder

# Set working directory
WORKDIR /app

COPY . .

RUN npm install -g pnpm

RUN cd front && pnpm install

RUN cd front && pnpm build

# Stage 2: Build the backend
FROM golang:1.23.4-alpine AS backend-builder

WORKDIR /app

COPY . .

# Copy backend source code
COPY --from=frontend-builder /app/server/static /app/server/static

# Build the Go application
RUN cd server && go mod download

RUN cd server && go build -o uchinoko main.go

# Stage 3: Create final image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy built frontend and backend
COPY --from=backend-builder /app/server/uchinoko /app/uchinoko

# Expose port
EXPOSE 8080

# Command to run the application
ENTRYPOINT ["./uchinoko"]