# Etapa 1: Construcción del binario
FROM golang:alpine AS builder

# Establece el directorio de trabajo dentro del contenedor
WORKDIR /worker

COPY . .

RUN go mod download

WORKDIR /worker
# Compila el binario en modo release
RUN go build -o worker .

# Etapa 2: Contenedor final (más ligero)
FROM alpine:latest

# Instala las dependencias necesarias
RUN apk add --no-cache curl && curl -fsSL https://download.docker.com/linux/static/stable/x86_64/docker-24.0.7.tgz | tar xz && mv docker/docker /usr/local/bin/ && chmod +x /usr/local/bin/docker && rm -rf docker

# Copia el binario desde la etapa de construcción
COPY --from=builder /worker/worker /usr/local/bin/worker

# Comando para ejecutar el servidor
CMD ["worker"]