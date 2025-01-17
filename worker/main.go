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
	// URL del servidor NATS (suponemos que está corriendo en la misma red o máquina)
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

	// Acceder al KV Store (asegúrate de que el bucket existe)
	/*kv, err := js.KeyValue("messages_store")
	if err != nil {
		log.Printf("Bucket 'messages_store' no encontrado, creándolo...")
		// Crear el bucket si no existe
		kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket: "messages_store",
		})
		if err != nil {
			log.Fatalf("Error al crear el bucket 'messages_store': %v", err)
		}
	} else if err != nil {
		log.Fatalf("Error al acceder al KV Store: %v")
	}*/

	// Suscribirse a la cola "queue.messages" para recibir las activaciones de funciones
	/*sub, err := nc.SubscribeSync("queue.messages")
	if err != nil {
		log.Fatalf("Error al suscribirse a la cola: %v", err)
	}*/

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
			Retention: nats.WorkQueuePolicy,    // Retención de mensajes basada en work queues
			MaxMsgs:   -1,                      // Ilimitado número de mensajes
			MaxBytes:  -1,                      // Ilimitado tamaño del stream
			MaxAge:    0,                       // Ilimitado tiempo de retención
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
			//js.Publish(ctx,"results."+requestId[1], []byte(fmt.Sprintf("Error para %s: %s", username, err.Error())))
			if msg.Reply != "" {
				nc.Publish(msg.Reply, []byte(fmt.Sprintf("Error: %v", err)))
			}
			return
		}

		// Guardar la solicitud en el KV Store
		/*kvKey := fmt.Sprintf("activation_%d", time.Now().UnixNano())
		_, err = kv.Put(kvKey, msg.Data)
		if err != nil {
			log.Printf("Error al guardar en el KV Store: %v", err)
		} else {
			log.Printf("Petición guardada en el KV Store con clave: %s", kvKey)
		}*/

		// Publicar el resultado de vuelta en el API Server
		if msg.Reply != "" {
			err = nc.Publish(msg.Reply, []byte(result))
			if err != nil {
				log.Printf("Error al enviar la respuesta: %v", err)
			} else {
				log.Printf("Resultado enviado: %s", result)
			}
		}

		msg.Ack() // Acknowledge message when processed
	})
	defer sub.Unsubscribe()

	if err != nil {
		log.Fatalf("Error subscribing to messages: %v", err)
	}

	log.Println("Worker esperando activaciones de funciones...")

	// Leer y procesar los mensajes
	/*for {
		// Esperar por el mensaje
		msg, err := sub.NextMsg(10 * time.Second) // Esperar hasta 10 segundos por un mensaje
		if err != nil {
			log.Printf("Timeout: No se recibió ningún mensaje en los últimos 10 segundos.")
			continue
		}

		// Mostrar el mensaje recibido
		log.Printf("Mensaje recibido: %s", string(msg.Data))

		username, functionName, param, err := SplitFunctionParam(string(msg.Data))
		if err != nil {
			log.Printf("Formato de mensaje inválido: %s", msg.Data)
			continue
		}

		// Procesar la función
		result, err := processFunction(workerMsgsId, functionName, param)
		if err != nil {
			log.Printf("Error ejecutando función para %s: %s\n", username, err.Error())
			//js.Publish(ctx,"results."+requestId[1], []byte(fmt.Sprintf("Error para %s: %s", username, err.Error())))
			nc.Publish(msg.Reply, []byte(fmt.Sprintf("Error: %v", err)))
			//return
			continue
		}

		// Guardar la solicitud en el KV Store
		kvKey := fmt.Sprintf("activation_%d", time.Now().UnixNano())
		_, err = kv.Put(kvKey, msg.Data)
		if err != nil {
			log.Printf("Error al guardar en el KV Store: %v", err)
		} else {
			log.Printf("Petición guardada en el KV Store con clave: %s", kvKey)
		}

		// Publicar el resultado de vuelta en el API Server
		err = nc.Publish(msg.Reply, []byte(result))
		if err != nil {
			log.Printf("Error al enviar la respuesta: %v", err)
		} else {
			log.Printf("Resultado enviado: %s", result)
		}
	}*/

	// Esperar señales de terminación (Ctrl+C, etc.)
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

// ejecuta un contenedor con el nombre de la función y un parámetro
/*func executeFunction(functionName, param string) (string, error) {
	ctx := context.Background()

	// Crear cliente Docker
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Descargar la imagen Docker si no está disponible localmente
	log.Printf("Descargando la imagen Docker: %s", functionName)
	_, err = cli.ImagePull(ctx, functionName, types.ImagePullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pull Docker image: %w", err)
	}

	// Crear y configurar el contenedor
	containerConfig := &container.Config{
		Image: functionName,
		Cmd:   []string{param}, // Pasar el parámetro como argumento
	}
	resp, err := cli.ContainerCreate(ctx, containerConfig, nil, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Iniciar el contenedor
	log.Printf("Iniciando contenedor: %s", resp.ID)
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	// Capturar la salida estándar del contenedor
	log.Printf("Obteniendo salida del contenedor: %s", resp.ID)
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve container logs: %w", err)
	}
	defer out.Close()

	// Leer y devolver el resultado
	result, err := io.ReadAll(out)
	if err != nil {
		return "", fmt.Errorf("failed to read container output: %w", err)
	}

	// Eliminar el contenedor
	log.Printf("Eliminando contenedor: %s", resp.ID)
	if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
		log.Printf("Warning: failed to remove container: %v", err)
	}

	return string(result), nil
}*/

// RunContainer ejecuta un contenedor Docker con una imagen y un parámetro
func RunContainer(image string, param string) ([]byte, error) {
	// Comando para ejecutar el contenedor
	cmd := exec.Command("docker", "run", "--rm", image, param)

	// Capturar la salida del contenedor
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error ejecutando el contenedor: %v\nSalida: %s", err, string(output))
	}

	fmt.Println("Salida del contenedor:", string(output))
	return output, nil
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
