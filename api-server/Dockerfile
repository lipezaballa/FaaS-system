# Etapa 1: Construcción del binario
FROM golang:alpine AS builder

# Establece el directorio de trabajo dentro del contenedor
WORKDIR /api-server

COPY . .

RUN go mod download

WORKDIR /api-server
# Compila el binario en modo release
RUN go build -o api-server .

# Etapa 2: Contenedor final (más ligero)
FROM alpine:latest

# Copia el binario desde la etapa de construcción
COPY --from=builder /api-server/api-server /usr/local/bin/api-server

# Exponer el puerto en el que corre el API server
EXPOSE 8080

# Comando para ejecutar el servidor
CMD ["api-server"]