package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	M int `json:"bits"`
}

func mysha(key string) int {
	m, err := ReadFromConfig()
	if err != nil {
		log.Fatal("Errore nella lettura dal file config.json: ", err)
	}

	//hash a partire da SHA-1 in modo che il valore di ritorno sia nel range (0, Numero di risorse nella rete)
	h := sha1.New()
	h.Write([]byte(key))
	res := byte((1 << m) - 1)
	hashedKey := h.Sum(nil)
	for i := 0; i < len(hashedKey); i++ {
		res = res ^ (hashedKey[i] % byte(1<<m))
	}
	return int(res)
}

func verifyIfNodeIsAlive(nodeId int, nodeIp string) int {
	var reply int
	arg := new(Arg)
	arg.Id = nodeId
	arg.Value = nodeIp
	client, err := rpc.DialHTTP("tcp", registryIpAddr)
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}

	err = client.Call("Registry.IsNodeAlive", arg, &reply)
	if err != nil {
		log.Fatal("Client invocation error: ", err)
		return -1
	}

	client.Close()
	return 0
}

func getMyIp() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Errore nell'ottenere l'hostname: ", err)
	}

	port := os.Getenv("NODE_PORT")

	addr, err := net.LookupHost(hostname)
	if err != nil {
		log.Fatal("Errore nell'ottenere l'indirizzo ip dell'host: ", err)
	}

	ipAddr := strings.Trim(addr[0], "[]")
	nodeIpPort := fmt.Sprintf("%s:%s", ipAddr, port)
	return nodeIpPort
}

func ReadFromConfig() (int, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return 0, err
	}
	filePath := filepath.Join(currentDir, "config.json")

	file, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return 0, err
	}

	return config.M, nil
}
