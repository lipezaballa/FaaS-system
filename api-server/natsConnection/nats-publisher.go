package natsConnection

import (
	"errors"
	"fmt"
	"log"
	"os"
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

	streamName := "messages_worker"
	_, err = js.StreamInfo(streamName)
	if err != nil {
		// El stream no existe, por lo tanto, lo creamos
		log.Printf("El stream '%s' no existe. Creando el stream...", streamName)

		// Configuración para crear el stream (puedes ajustarlo según tus necesidades)
		streamConfig := &nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{channel},
			Retention: nats.WorkQueuePolicy,    // Retención de mensajes basada en work queues
			MaxMsgs:   -1,                      // Ilimitado número de mensajes
			MaxBytes:  -1,                      // Ilimitado tamaño del stream
			MaxAge:    0,                       // Ilimitado tiempo de retención
			Storage:   nats.MemoryStorage,      // Almacenar en memoria (puedes usar FileStorage)
		}

		// Intentamos crear el stream
		_, err = js.AddStream(streamConfig)
		if err != nil {
			log.Fatalf("Error creando el stream: %v", err)
		}

		log.Printf("Stream '%s' creado exitosamente", channel)
	} else {
		log.Printf("El stream '%s' ya existe", channel)
	}



	// Crear un stream si no existe
	/*_, err = js.AddStream(&nats.StreamConfig{
		Name:     "messages_stream",  // Nombre del stream
		Subjects: []string{channel},  // El subject que usará el worker
		Storage:  nats.MemoryStorage, // Almacenamiento en memoria
	})
	if err != nil {
		log.Fatalf("Error creando el stream: %v", err)
	}*/

	natsConn.Kv = kv
	natsConn.Channel = channel
	return natsConn, nil
}

func SendRequest(msg string) (*nats.Msg, error) {
	// Enviar mensaje y esperar respuesta
	if (natsConn != nil && natsConn.Nc != nil) {
		resp, err := natsConn.Nc.Request("queue.messages.worker", []byte(msg), 5*time.Second)
		if err != nil {
			fmt.Fprintf(os.Stderr, "No se recibió respuesta: %v", err)
			return nil, err
		} else {
			log.Printf("Respuesta recibida: %s\n", string(resp.Data))
			return resp, nil
		}
	} else {
			fmt.Fprintf(os.Stderr, "No existe la conexión")
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