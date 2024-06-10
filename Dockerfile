FROM node:20-alpine AS frontend-builder

# Set working directory
WORKDIR /app

COPY . .

RUN npm install -g pnpm

RUN cd front && pnpm install && pnpm build

# Stage 2: Build the backend
FROM golang:1.22-alpine AS backend-builder

COPY . .

# Set working directory
WORKDIR /app/server

# Copy backend source code
COPY --from=frontend-builder /app/server/static /app/server/static

# Build the Go application
RUN go build -o uchinoko .

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