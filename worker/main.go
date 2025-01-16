package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	// URL del servidor NATS (suponemos que está corriendo en la misma red o máquina)
	natsURL := "nats://nats_server:4222"
	log.Println("Conectando al servidor NATS...")

	// Conectar a NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Error al conectar a NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Conexión a NATS exitosa.")

	// Conectar a JetStream
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Error al inicializar JetStream: %v", err)
	}

	// Acceder al KV Store (asegúrate de que el bucket existe)
	kv, err := js.KeyValue("messages_store")
	if err != nil {
		log.Fatalf("Error al acceder al KV Store: %v", err)
	}

	// Suscribirse a la cola "queue.messages" para recibir los mensajes
	sub, err := nc.SubscribeSync("queue.messages")
	if err != nil {
		log.Fatalf("Error al suscribirse a la cola: %v", err)
	}

	log.Println("Worker esperando mensajes...")

	// Leer y procesar los mensajes
	for {
		// Esperar por el mensaje
		msg, err := sub.NextMsg(10 * time.Second) // Esperar hasta 10 segundos por un mensaje
		if err != nil {
			log.Printf("Error al recibir mensaje, si no se recibe en 10 segs salta timeout: %v", err)
			continue
		}

		// Mostrar el mensaje recibido
		log.Printf("Mensaje recibido: %s", string(msg.Data))

		// Leer o escribir en el KV Store (ejemplo: escribir el mensaje recibido con un timestamp)
		//timestamp := time.Now().Format(time.RFC3339)
		kvKey := "key11"
		_,err = kv.Put(kvKey, msg.Data)
		if err != nil {
			log.Printf("Error al guardar en el KV Store: %v", err)
		} else {
			log.Printf("Petición guardada en el KV Store con clave: %s", kvKey)
		}

		// Enviar una respuesta de vuelta al API Server
		//responseMessage := fmt.Sprintf("Worker procesó el mensaje: %s", string(msg.Data))
		responseMessage := fmt.Sprintf("Resultado 1")
		err = nc.Publish(msg.Reply, []byte(responseMessage))
		if err != nil {
			log.Printf("Error al enviar la respuesta: %v", err)
		} else {
			log.Printf("Respuesta enviada: %s", responseMessage)
		}
	}
}