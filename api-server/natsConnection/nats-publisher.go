package natsConnection

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lipezaballa/FaaS-system/shared"
	"github.com/nats-io/nats.go"
)

func InitJetStream(nc *nats.Conn, channel string) (*shared.NatsConnection, error)  {
	natsConnetion := &shared.NatsConnection{
		Nc: nc,
	}

	log.Println("iniciar JetStream")
	// Key-Value Store para historial de mensajes
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Error al inicializar JetStream: %v", err)
		return nil, err
	}
	natsConnetion.Js = &js

	log.Println("creando database messages_store")
	kv, err := js.KeyValue("messages_store")
	if err != nil {
		log.Printf("Bucket 'messages_store' no encontrado, creándolo...")
		// Crear el bucket si no existe
		kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket: "messages_store",
		})
		if err != nil {
			log.Fatalf("Error al crear el KV Store: %v", err)
			return nil, err
		}
	}
	natsConnetion.Kv = kv
	natsConnetion.Channel = channel
	return natsConnetion, nil
}

func SendRequest(natsConnection *shared.NatsConnection, msg string) (*nats.Msg, error) {
	log.Println("enviando request")
	// Enviar mensaje y esperar respuesta
	if (natsConnection.Nc != nil) {
		resp, err := natsConnection.Nc.Request("queue.messages", []byte(msg), 2*time.Second)
		if err != nil {
			log.Printf("No se recibió respuesta: %v", err)
			return nil, err
		} else {
			fmt.Printf("Respuesta recibida: %s\n", string(resp.Data))
			return resp, nil
		}
	} else {
			log.Printf("No existe la conexión")
			err := errors.New("No existe la conexión")
			return nil, err
	}
}

func StoreInKv(natsConnection *shared.NatsConnection, msg string) error {
	log.Println("Guardar en kv store")
	if (natsConnection.Kv != nil) {
		// Guardar en KV Store
		_, err := natsConnection.Kv.Put("hola", []byte(msg))
		if err != nil {
			log.Fatalf("Error al guardar en KV Store: %v", err)
			return err
		}

		fmt.Println("Mensaje enviado y almacenado:", msg)
		return nil
	} else {
		log.Printf("No existe KV Store")
		err := errors.New("No existe KV Store")
		return err
	}
}

func GetValues(natsConnection *shared.NatsConnection) error {

	if (natsConnection.Kv != nil) {
	keys, err := natsConnection.Kv.Keys()
	if err != nil {
		log.Fatalf("Error al obtener claves de KV Store: %v", err)
	}

	for _, key := range keys {
		entry, err := natsConnection.Kv.Get(key) //FIXME
		if err == nil {
			fmt.Printf("Histórico: %s\n", string(entry.Value()))
		}

	}
	return nil
	} else {
		log.Printf("No existe KV Store")
		err := errors.New("No existe KV Store")
		return err
	}
}

func testNATS() { //FIXME quit
	//natsURL := os.Getenv("NATS_URL")
	natsURL := "nats://nats_server:4222"
	log.Println("ejecutando testNATS")
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Error al conectar a NATS: %v", err)
	}
	defer nc.Close()

	log.Println("iniciar JetStream")
	// Key-Value Store para historial de mensajes
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Error al inicializar JetStream: %v", err)
	}

	log.Println("creando database messages_store")
	kv, err := js.KeyValue("messages_store")
	if err != nil {
		log.Printf("Bucket 'messages_store' no encontrado, creándolo...")
		// Crear el bucket si no existe
		kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket: "messages_store",
		})
		if err != nil {
			log.Fatalf("Error al crear el KV Store: %v", err)
		}
	}

	//for {
	timestamp := time.Now().Format(time.RFC3339)
	message := fmt.Sprintf("[%s] Usuario1: Hola desde el Publisher", timestamp)

	log.Println("enviando request")
	// Enviar mensaje y esperar respuesta
	resp, err := nc.Request("queue.messages", []byte(message), 2*time.Second)
	if err != nil {
		log.Printf("No se recibió respuesta: %v", err)
	} else {
		fmt.Printf("Respuesta recibida: %s\n", string(resp.Data))
	}

	log.Println("Guardar en kv store")
	// Guardar en KV Store
	_, err = kv.Put("hola", []byte(message))
	if err != nil {
		log.Fatalf("Error al guardar en KV Store: %v", err)
	}

	fmt.Println("Mensaje enviado y almacenado:", message)
	time.Sleep(5 * time.Second)

	keys, err := kv.Keys()
	if err != nil {
		log.Fatalf("Error al obtener claves de KV Store: %v", err)
	}

	for _, key := range keys {
		entry, err := kv.Get(key)
		if err == nil {
			fmt.Printf("Histórico: %s\n", string(entry.Value()))
		}

	}
	//}
}