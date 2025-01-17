package natsConnection

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lipezaballa/FaaS-system/shared"
	"github.com/nats-io/nats.go"
)

var natsConn *shared.NatsConnection

func InitJetStream(nc *nats.Conn, channel string) (*shared.NatsConnection, error)  {
	natsConn= &shared.NatsConnection{
		Nc: nc,
	}

	log.Println("iniciar JetStream")
	// Key-Value Store para historial de mensajes
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Error al inicializar JetStream: %v", err)
		return nil, err
	}
	natsConn.Js = &js

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
	natsConn.Kv = kv
	natsConn.Channel = channel
	return natsConn, nil
}

func SendRequest(msg string) (*nats.Msg, error) {
	log.Println("enviando request")
	// Enviar mensaje y esperar respuesta
	if (natsConn != nil && natsConn.Nc != nil) {
		resp, err := natsConn.Nc.Request("queue.messages", []byte(msg), 2*time.Second)
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

func StoreInKv(natsConnection *shared.NatsConnection, msg string) error { //FIXME quit
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

func StoreFunction(username string, functionName string, functionImage string) error {
	log.Println("Guardar en kv store")
	if (natsConn != nil && natsConn.Kv != nil) {
		// Guardar en KV Store
		key := fmt.Sprintf("users/%s/functions/%s", username, functionName)
		_, err := natsConn.Kv.Put(key, []byte(functionImage))
		if err != nil {
			log.Fatalf("Error al guardar en KV Store: %v", err)
			return err
		}

		fmt.Println("Mensaje enviado y almacenado:", key)
		return nil
	} else {
		log.Printf("No existe KV Store")
		err := errors.New("No existe KV Store")
		return err
	}
}

func StoreUser(username string, password string) error {
	log.Println("Guardar en kv store")
	if (natsConn != nil && natsConn.Kv != nil) {
		// Guardar en KV Store
		//key := fmt.Sprintf("users/%s/functions/%s", username, functionName)
		_, err := natsConn.Kv.Put(username, []byte(password))
		if err != nil {
			log.Fatalf("Error al guardar en KV Store: %v", err)
			return err
		}

		fmt.Println("Mensaje enviado y almacenado:", username)
		return nil
	} else {
		log.Printf("No existe KV Store")
		err := errors.New("No existe KV Store")
		return err
	}
}

func PrintValues() error {

	if (natsConn != nil && natsConn.Kv != nil) {
		keys, err := natsConn.Kv.Keys()
		if err != nil {
			log.Fatalf("Error al obtener claves de KV Store: %v", err)
		}

		for _, key := range keys {
			entry, err := natsConn.Kv.Get(key) //FIXME
			if err == nil {
				fmt.Printf("Database: Key: %s, Value: %s\n", key, string(entry.Value()))
			}

		}
		return nil
	} else {
		log.Printf("No existe KV Store")
		err := errors.New("No existe KV Store")
		return err
	}
}

func GetValue(key string) (nats.KeyValueEntry,bool) {
	if (natsConn != nil && natsConn.Kv != nil) {

		entry, err := natsConn.Kv.Get(key)
		if err != nil {
			log.Printf("Error al obtener la clave '%s': %v", key, err)
			return nil, false
		}
		return entry, true

	} else {
		log.Printf("No existe KV Store")
		return nil, false
	}
}

func DeleteKeyFromKV(key string) error {
    log.Printf("Eliminando clave '%s' del KV Store...", key)

	if (natsConn != nil && natsConn.Kv != nil) {

		err := natsConn.Kv.Delete(key)
		if err != nil {
			log.Printf("Error al eliminar la clave '%s': %v", key, err)
			return err
		}

		log.Printf("Clave '%s' eliminada con éxito", key)
		return nil

	} else {
		log.Printf("No existe KV Store")
		err := errors.New("No existe KV Store")
		return err
	}
}

func DeleteAllKeysFromKV() error {
	if (natsConn != nil && natsConn.Kv != nil) {
		keys, err := natsConn.Kv.Keys()
		if err != nil {
			log.Fatalf("Error al obtener claves de KV Store: %v", err)
		}

		for _, key := range keys {
			err := natsConn.Kv.Delete(key) //FIXME
			if err != nil {
				log.Printf("Error al eliminar la clave '%s': %v", key, err)
				return err
			}
	
			log.Printf("Clave '%s' eliminada con éxito", key)
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