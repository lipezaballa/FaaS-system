package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

func main() {
	natsURL := "nats://nats_server:4222"
	log.Println("Conectando al servidor NATS...")

	workerMsgsId := "worker_" + uuid.NewString()

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

	// Verificamos si el stream existe
	streamName := "messages_worker"
	_, err = js.StreamInfo(streamName)
	if err != nil {
		// El stream no existe, por lo tanto, lo creamos
		log.Printf("El stream '%s' no existe. Creando el stream...", streamName)

		// Configuración para crear el stream (puedes ajustarlo según tus necesidades)
		streamConfig := &nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{"queue.messages.worker"},
			Retention: nats.WorkQueuePolicy,    
			MaxMsgs:   -1,                      
			MaxBytes:  -1,                      
			MaxAge:    0,                       
			Storage:   nats.MemoryStorage, 
		}

		// Intentamos crear el stream
		_, err = js.AddStream(streamConfig)
		if err != nil {
			log.Fatalf("Error creando el stream: %v", err)
		}

		log.Printf("Stream '%s' creado exitosamente", streamName)
	} else {
		log.Printf("El stream '%s' ya existe", streamName)
	}

	// Subscribe to the queue messages and use the same queue group for all workers
	sub, err := js.QueueSubscribe("queue.messages.worker", "worker_group", func(msg *nats.Msg) {
		log.Printf("Worker received message: %s", string(msg.Data))
		// Process the message here (run docker command, etc.)
		username, functionName, param, err := SplitFunctionParam(string(msg.Data))
		if err != nil {
			log.Printf("Formato de mensaje inválido: %s", msg.Data)
			return
		}

		// Procesar la función
		result, err := processFunction(workerMsgsId, functionName, param)
		if err != nil {
			log.Printf("Error ejecutando función para %s: %s\n", username, err.Error())

			if msg.Reply != "" {
				nc.Publish(msg.Reply, []byte(fmt.Sprintf("Error: %v", err)))
			}
			return
		}

		// Publicar el resultado de vuelta en el API Server
		if msg.Reply != "" {
			err = nc.Publish(msg.Reply, []byte(result))
			if err != nil {
				log.Printf("Error al enviar la respuesta: %v", err)
			} else {
				log.Printf("Resultado enviado: %s", result)
			}
		}

		msg.Ack()
	})
	defer sub.Unsubscribe()

	if err != nil {
		log.Fatalf("Error subscribing to messages: %v", err)
	}

	log.Println("Worker esperando activaciones de funciones...")

	// Esperar señales de terminación
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan // Bloquear hasta recibir una señal de terminación
	log.Println("Worker terminado.")
}

// SplitFunctionParam divide una cadena en función y parámetro
func SplitFunctionParam(input string) (string, string, string, error) {
	parts := strings.Split(input, "|")

	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("el formato es incorrecto, se esperaba 'functionname|param'")
	}

	return parts[0], parts[1], parts[2], nil
}

func processFunction(workerMsgsId, functionName, parameter string) (string, error) {
	// Simulación de ejecución con Docker
	log.Printf("[%s] PROCESANDO la función %s", workerMsgsId, functionName)
	cmd := exec.Command("docker", "run", "--rm", functionName, parameter)
	fmt.Println("Ejecutando comando:", strings.Join(cmd.Args, " "))
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error ejecutando la función: %s, %s", stderr.String(), err)
	}
	return out.String(), nil
}
