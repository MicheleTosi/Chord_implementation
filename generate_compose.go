package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func getNumContainersFromConfig() int {
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Errore nella lettura del file di configurazione: %v", err)
	}

	// Analizza il file JSON
	var configData map[string]interface{}
	if err := json.Unmarshal(configFile, &configData); err != nil {
		log.Fatalf("Errore nel parsing del file di configurazione JSON: %v", err)
	}

	nodes, ok := configData["nodes"].(float64)
	if !ok {
		log.Fatalf("Il campo 'nodes' nel file di configurazione deve essere un numero.")
	}

	return int(nodes)
}

func generateDockerCompose(numContainers int) {
	// Apertura del file docker-compose.yml in modalità scrittura
	file, err := os.Create("docker-compose.yml")
	if err != nil {
		log.Fatalf("Errore nell'apertura del file docker-compose.yml: %v", err)
	}
	defer file.Close()

	// Scrittura del contenuto nel file
	file.WriteString("version: \"3\"\n\n")
	file.WriteString("services:\n")
	file.WriteString("  registry:\n")
	file.WriteString("    build:\n")
	file.WriteString("      context: ./ #Si parte dalla root del progetto\n")
	file.WriteString("      dockerfile: ./DockerFiles/registry/Dockerfile #Path del docker file del registry\n")
	file.WriteString("    ports:\n")
	file.WriteString("      - \"1234:1234\" #Mappa la porta 1234 del server del container a quella dell'host\n")
	file.WriteString("    networks:\n")
	file.WriteString("      - my_network\n\n")

	//crea i servizi node con nomi e porte distinte
	for i := 1; i <= numContainers; i++ {
		file.WriteString(fmt.Sprintf("  node%d:\n", i))
		file.WriteString("    build:\n")
		file.WriteString("      context: ./\n")
		file.WriteString("      dockerfile: ./DockerFiles/node/Dockerfile\n")
		file.WriteString(fmt.Sprintf("    ports:\n      - \"%d:%d\" #Mappa la porta %d del server con quella del container\n", 8000+i, 8000+i, 8000+1))
		file.WriteString(fmt.Sprintf("    environment:\n      - NODE_PORT=%d #Porta esposta dal nodo\n", 8000+i))
		file.WriteString("    networks:\n")
		file.WriteString("      - my_network\n")
		file.WriteString("    depends_on:\n")
		file.WriteString("      - registry\n\n")
	}

	file.WriteString("networks:\n")
	file.WriteString("  my_network:\n")
	file.WriteString("    driver: bridge\n")
}

func main() {
	numContainers := getNumContainersFromConfig()
	generateDockerCompose(numContainers)
}
