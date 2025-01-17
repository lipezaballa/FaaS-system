# Etapa 1: Construcción del binario
FROM golang:alpine AS builder

# Establece el directorio de trabajo dentro del contenedor
WORKDIR /worker

#COPY go.work go.work.sum ./
#COPY ./api-server/go.mod ./api-server/go.sum ./
#COPY ./reverse-proxy/go.mod ./reverse-proxy/go.sum ./
#COPY ./shared/go.mod ./
COPY . .

RUN go mod download

# Copia los archivos del proyecto al contenedor
# COPY api-server/ .
#COPY . .

#ENV GOWORK=on
# Descarga las dependencias del módulo
# RUN go mod tidy
WORKDIR /worker
# Compila el binario en modo release
RUN go build -o worker .

# Etapa 2: Contenedor final (más ligero)
FROM alpine:latest

# Instala las dependencias necesarias
RUN apk add --no-cache curl && curl -fsSL https://download.docker.com/linux/static/stable/x86_64/docker-24.0.7.tgz | tar xz && mv docker/docker /usr/local/bin/ && chmod +x /usr/local/bin/docker && rm -rf docker

# Establece el directorio de trabajo
# WORKDIR /root/

# Copia el binario desde la etapa de construcción
COPY --from=builder /worker/worker /usr/local/bin/worker

# Exponer el puerto en el que corre el API server
# EXPOSE 8080

# Comando para ejecutar el servidor
CMD ["worker"]