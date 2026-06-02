# Estágio 1: construir o frontend React
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Estágio 2: construir o backend Go
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copia o frontend compilado para o local onde o embed espera
COPY --from=frontend-builder /app/frontend/dist ./cmd/server/web/dist
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

# Estágio 3: imagem final enxuta
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=backend-builder /server /server
EXPOSE 8080
CMD ["/server"]