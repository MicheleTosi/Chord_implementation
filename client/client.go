package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
)

type Arg struct {
	Id    int
	Value string
	Type  string
}

var registryIpAddr string = "0.0.0.0:1234"

//client
//quando il client vuole contattare un nodo contatta il registry (porta fissa 1234) che gli ritorna un nodo in modo RoundRobin
//il client viene utilizzato per:
//aggiungere, rimuovere, cercare risorsa
//rimuovere nodo dalla rete

func main() {

	arg := new(Arg)

	var reply string

	var b int

	for {
		fmt.Println("Selezionare una delle seguenti opzioni:")
		fmt.Println("1. Inserire un nuovo oggetto")
		fmt.Println("2. Rimuovere un oggetto")
		fmt.Println("3. Cercare un oggetto")
		fmt.Println("4. Rimuovere un nodo")
		fmt.Print("\nScegliere un'opzione e premere invio: ")
		fmt.Scanln(&b)
		switch int(b) {
		case 1:
			fmt.Print("Digita l'oggetto da inserire: ")

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			arg.Value = scanner.Text()

			client, err := rpc.DialHTTP("tcp", registryIpAddr)
			if err != nil {
				log.Fatal("Client connection error: ", err)
			}

			err = client.Call("Registry.ReturnChordNode", arg, &reply)
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}

			client.Close()

			reply = "0.0.0.0" + reply[len(reply)-5:]
			client, err = rpc.DialHTTP("tcp", reply)
			if err != nil {
				log.Fatal("Client connection error: ", err)
			}

			err = client.Call("ChordNode.AddObject", arg, &reply)
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}

			client.Close()
			fmt.Println(reply)

		case 2:
			fmt.Print("Digita l'id dell'oggetto da rimuovere: ")

			fmt.Scanln(&arg.Id)

			client, err := rpc.DialHTTP("tcp", registryIpAddr)
			if err != nil {
				log.Fatal("Client connection error: ", err)
			}

			err = client.Call("Registry.ReturnChordNode", arg, &reply)
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}
			client.Close()

			reply = "0.0.0.0" + reply[len(reply)-5:]

			client, err = rpc.DialHTTP("tcp", reply)
			if err != nil {
				log.Fatal("Client connection error2: ", err)
			}

			err = client.Call("ChordNode.RemoveObject", arg, &reply)
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}
			client.Close()

			fmt.Println(reply)

		case 3:
			fmt.Print("Digita l'id dell'oggetto da cercare: ")

			fmt.Scanln(&arg.Id)

			rep := new(Arg)

			client, err := rpc.DialHTTP("tcp", registryIpAddr)
			if err != nil {
				log.Fatal("Client connection error: ", err)
			}

			err = client.Call("Registry.ReturnChordNode", arg, &reply)
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}
			client.Close()

			reply = "0.0.0.0" + reply[len(reply)-5:]

			client, err = rpc.DialHTTP("tcp", reply)
			if err != nil {
				log.Fatal("Client connection error2: ", err)
			}

			err = client.Call("ChordNode.SearchObject", arg, &rep)
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}
			client.Close()

			if rep.Value == "" {
				fmt.Println("L'oggetto cercato non è presente.")
			} else {
				str := fmt.Sprintf("L'oggetto con id %d è %s.", arg.Id, rep.Value)
				fmt.Println(str)
			}

		case 4:
			client, err := rpc.DialHTTP("tcp", registryIpAddr)
			if err != nil {
				log.Fatal("Client connection error: ", err)
			}

			err = client.Call("Registry.GetNodes", arg, &reply)
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}
			client.Close()
			fmt.Println(reply)

			fmt.Print("Digita l'id del nodo da rimuovere: ")

			fmt.Scanln(&arg.Id)

			client, err = rpc.DialHTTP("tcp", registryIpAddr)
			if err != nil {
				log.Fatal("Client connection error: ", err)
			}

			err = client.Call("Registry.RemoveNode", arg, &reply)
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}
			client.Close()

			fmt.Println(reply)

		default:
			fmt.Println("Devi selezionare una delle scelte digitando un numero tra 1 e 4")
		}

		fmt.Println("Premere INVIO per continuare...")
		fmt.Scanln()               //Attende che l'utente prema INVIO
		fmt.Print("\033[H\033[2J") //pulisce la console
	}

}
