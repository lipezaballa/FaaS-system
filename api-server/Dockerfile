# Etapa 1: Construcción del binario
FROM golang:alpine AS builder

# Establece el directorio de trabajo dentro del contenedor
WORKDIR /api-server

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
WORKDIR /api-server
# Compila el binario en modo release
RUN go build -o api-server .

# Etapa 2: Contenedor final (más ligero)
FROM alpine:latest

# Instala las dependencias necesarias
# RUN apk --no-cache add ca-certificates

# Establece el directorio de trabajo
# WORKDIR /root/

# Copia el binario desde la etapa de construcción
COPY --from=builder /api-server/api-server /usr/local/bin/api-server

# Exponer el puerto en el que corre el API server
EXPOSE 8080

# Comando para ejecutar el servidor
CMD ["api-server"]