package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sort"
	"time"
)

// ogni nodo mantiene il suo id e ipaddr (ip+port), fingerTable e risorse che gestisce
// mantiene inoltre successore e predecessore in modo da conoscere i nodi vicini per gestire ingressi e uscite
type Node struct {
	Id          int
	Predecessor string
	Successor   string
	Ipaddr      string
	FingerTable map[int]string
	Objects     map[int]string
}

type Arg struct {
	Id    int
	Value string
	Type  string //'s' se è ricerca, 'f' se è per costruire la fingerTable, 'a' per aggiungere o 'd' cancellare l'oggetto
}

type Neighbors struct {
	Predecessor string
	Successor   string
}

type ChordNode string

var node *Node

var registryIpAddr string = "registry:1234"
var exitChan = make(chan struct{})

/*myKeys
*viene invocata dal nodo per sapere se l'id di interesse è gestito da lui
 */
func myKeys(currIp string, idPredecessor int, objId int) bool { //funzione che torna true se l'id dell'oggetto è del nodo che invoca la funz altrimenti torna false
	id := mysha(currIp)
	//se il predecessore del nodo corrente è il nodo stesso c'è solo lui nella rete e gestisce tutti i nodi
	if currIp == node.Predecessor {
		return true
	}
	//se l'id dell'oggetto è compreso tra l'id del nodo corrente e del predecessore, l'oggetto viene gestito dal nodo corrente
	if objId > idPredecessor && objId <= id {
		return true
	}
	//se il nodo corrente ha id più piccolo del predecessore bisogna controllare se l'id dell'oggetto è più grande di quello del predecessore o più piccolo del nodo corrente
	if id < idPredecessor && (objId <= id || objId > idPredecessor) {
		return true
	}
	return false //in tutti gli altri casi il nodo corrente non gestisce l'oggetto
}

func getPredAndSucc() {
	neighbors := new(Neighbors)
	arg := new(Arg)
	arg.Value = node.Ipaddr
	arg.Id = node.Id

	client, err := rpc.DialHTTP("tcp", registryIpAddr)
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}

	err = client.Call("Registry.Neighbors", arg, &neighbors)
	if err != nil {
		log.Fatal("Client invocation error: ", err)
	}

	client.Close()

	node.Successor = neighbors.Successor
	node.Predecessor = neighbors.Predecessor
}

/*enterRing
*viene invocata dal nodo per contattare successore e predecessore per notificarli del suo ingresso nella rete, in più il successore gli ritorna
*le risorse che il nuovo nodo dovrà gestire
 */
func enterRing() {
	if node.Ipaddr != node.Successor { //contatto il successore per recuperare l'ip del predecessore e le mie risorse se non coincide con me stesso
		arg := new(Arg)
		arg.Value = node.Ipaddr
		arg.Id = node.Id

		//contatto il successore per farmi restituire le chiavi e per fargli aggiornare il predecessore
		client, err := rpc.DialHTTP("tcp", node.Successor)
		if err != nil {
			log.Fatal("Client connection error: ", err)
		}

		err = client.Call("ChordNode.NotifySuccOfPredChange", arg, &node.Objects)
		if err != nil {
			log.Fatal("Client invocation error: ", err)
		}

		client.Close()

		//contatto il predecessore per fargli aggiornare il successore
		client, err = rpc.DialHTTP("tcp", node.Predecessor)
		if err != nil {
			log.Fatal("Client connection error: ", err)
		}

		err = client.Call("ChordNode.NotifyPredOfSuccChange", node, nil)
		if err != nil {
			log.Fatal("Client invocation error: ", err)
		}

		client.Close()
	}

}

func (t *ChordNode) NotifySuccOfPredChange(arg *Arg, reply *map[int]string) error {
	node.Predecessor = arg.Value
	for k := range node.Objects {
		if !myKeys(node.Ipaddr, arg.Id, k) {
			(*reply)[k] = node.Objects[k]
			delete(node.Objects, k)
		}
	}

	return nil
}

func (t *ChordNode) NotifyPredOfSuccChange(arg *Node, reply *string) error {
	node.Successor = arg.Ipaddr

	for k := range node.FingerTable {
		if myKeys(arg.Ipaddr, node.Id, k) {
			node.FingerTable[k] = arg.Ipaddr //aggiorno le entry della fingerTable del predecessore
		}
	}

	return nil
}

func updateFingerTable() {
	//TODO logica per aggiornare la fingerTable
	for {
		select {
		case <-exitChan:
			fmt.Println("Connessione interrotta correttamente.")
			os.Exit(0)
		default:
			fingerTable()
			printFingerTable()
			time.Sleep(15 * time.Second) //ogni 15 secondi aggiorno la fingerTable fino a che il nodo non esce dall'anello
		}

	}
}

func printFingerTable() {
	m, err := ReadFromConfig()
	if err != nil {
		log.Fatal("Errore nella lettura dal file config.json: ", err)
	}

	fmt.Printf("FT nodo %s\n", node.Ipaddr)
	for i := 0; i < m; i++ {
		k := (node.Id + (1 << i)) % (1 << m)
		fmt.Printf("FT[%d]=%s\n", k, node.FingerTable[k])
	}
}

func fingerTable() {
	//il nodo riceve la richiesta con id cercato, risponde con suo id e suo ip se è lui a gestire la risorsa altrimenti ritorna l'ip
	//del nodo da contattare per saperne di più ossia il primo nodo nella finger table con id minore di quello della risorsa cercata

	m, err := ReadFromConfig()
	if err != nil {
		log.Fatal("Errore nella lettura dal file config.json: ", err)
	}

	reply := new(Arg)
	arg := new(Arg) //argomento da passare alla funzione per conoscere il successore della risorsa (2^i)%N
	arg.Type = "f"
	node.FingerTable[node.Id+1] = node.Successor

	//TODO logica per costruire la fingerTable
	for i := 1; i < m; i++ { //TODO sostituire 32 con la variabile N numero di risorse per renderlo variabile
		//devo fare in modo che cerco "int(math.Pow(2, i)) % N" e ottengo ip e id (faccio chiamata al successore)
		//fingerTable[int(math.Pow(2, i))%N] = "127.0.0.1:5000" //modificare indirizzo in questo modo fingerTable statica
		//FIXIT fingerTable deve essere in modo tale che fingerTable[idNodo]=ipNodo

		arg.Id = (node.Id + 1<<i) % (1 << m)
		if myKeys(node.Ipaddr, mysha(node.Predecessor), arg.Id) {
			node.FingerTable[arg.Id] = node.Ipaddr
		} else {
			ipAddr := getIpByFingerTable(arg.Id)
			if ipAddr == node.Ipaddr {
				ipAddr = node.Successor
			}
			client, err := rpc.DialHTTP("tcp", ipAddr) //contatto l'ip ottenuto dalla fingerTable per conoscere l'ip del nodo che gestisce la risorsa desiderata

			if err != nil {
				result := verifyIfNodeIsAlive(mysha(ipAddr), ipAddr)
				if result != 0 {
					log.Fatal("Errore nodo Finger: il nodo da contattare è attivo, ma non riesco a instaurare una connessione.", err)
				} else {
					return
				}

			}

			err = client.Call("ChordNode.SearchObject", arg, &reply) //invoco la funzione finger del successore passandogli l'id della risorsa che mi interessa
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}

			client.Close()

			node.FingerTable[arg.Id] = reply.Value
		}
	}

}

func (t *ChordNode) SearchObject(arg *Arg, reply *Arg) error {
	id := arg.Id
	idPredecessor := mysha(node.Predecessor)

	if myKeys(node.Ipaddr, idPredecessor, id) {
		if arg.Type == "f" {
			reply.Value = node.Ipaddr
			return nil
		}
		if node.Objects[arg.Id] == "" { //se la risorsa cercata non c'è
			reply.Value = ""
			if arg.Type == "a" {
				node.Objects[arg.Id] = arg.Value
			}
		} else {
			reply.Value = node.Objects[arg.Id]
			reply.Id = node.Id
			if arg.Type == "d" {
				delete(node.Objects, id)
			}
		}
	} else {

		ipAddr := getIpByFingerTable(id)

		client, err := rpc.DialHTTP("tcp", ipAddr)
		if err != nil {
			result := verifyIfNodeIsAlive(mysha(ipAddr), ipAddr)

			if result != 0 {
				log.Fatal("Errore nodo Finger: il nodo da contattare è attivo, ma non riesco a instaurare una connessione.", err)
			} else {
				return nil
			}
		}

		err = client.Call("ChordNode.SearchObject", arg, &reply)
		if err != nil {
			log.Fatal("Client invocation error: ", err)
		}

		client.Close()
	}

	return nil
}

func getIpByFingerTable(id int) string {

	keys := make([]int, 0, len(node.FingerTable))
	for k := range node.FingerTable {
		keys = append(keys, k)
	}

	sort.Sort(sort.IntSlice(keys))

	for i, v := range keys {
		if id < v {
			if i == 0 {
				return node.FingerTable[keys[i]]
			} else {
				return node.FingerTable[keys[i-1]]
			}
		}
	}
	return node.FingerTable[keys[len(keys)-1]]
}

func (t *ChordNode) AddObject(arg *Arg, reply *string) error {
	//cerca l'ip del nodo che dovrebbe gestire il nuovo oggetto
	//lo contatta per aggiungere l'oggetto
	rep := new(Arg)
	arg.Id = mysha(arg.Value)
	arg.Type = "a"
	ipAddr := getIpByFingerTable(arg.Id)
	client, err := rpc.DialHTTP("tcp", ipAddr)
	if err != nil {
		result := verifyIfNodeIsAlive(mysha(ipAddr), ipAddr)
		if result != 0 {
			log.Fatal("Errore nodo Finger: il nodo da contattare è attivo, ma non riesco a instaurare una connessione.", err)
		} else {
			return nil
		}
	}

	err = client.Call("ChordNode.SearchObject", arg, rep)
	if err != nil {
		log.Fatal("Client invocation error: ", err)
	}

	client.Close()

	if rep.Value != "" {
		*reply = "L'oggetto " + arg.Value + " non può essere inserito in quanto esiste già un oggetto con lo stesso id."
	} else {
		*reply = fmt.Sprintf("Oggetto '%s' inserito con id %d.", arg.Value, mysha(arg.Value))
	}

	return nil
}

func (t *ChordNode) RemoveObject(arg *Arg, reply *string) error {
	//cerca l'ip del nodo che dovrebbe gestire il nuovo oggetto
	//lo contatta per aggiungere l'oggetto
	rep := new(Arg)
	arg.Type = "d"
	ipAddr := getIpByFingerTable(arg.Id)
	client, err := rpc.DialHTTP("tcp", ipAddr)
	if err != nil {
		result := verifyIfNodeIsAlive(mysha(ipAddr), ipAddr)
		if result != 0 {
			log.Fatal("Errore nodo Finger: il nodo da contattare è attivo, ma non riesco a instaurare una connessione.", err)
		} else {
			return nil
		}
	}

	err = client.Call("ChordNode.SearchObject", arg, rep)
	if err != nil {
		log.Fatal("Client invocation error: ", err)
	}

	client.Close()

	if rep.Value == "" {
		*reply = "L'oggetto da rimuovere non è presente."
	} else {
		*reply = fmt.Sprintf("L'oggetto '%s' con id %d è stato rimosso.", rep.Value, arg.Id)
	}
	return nil
}

func (t *ChordNode) ExitRing(arg *Arg, reply *string) error {
	var rep string
	client, err := rpc.DialHTTP("tcp", node.Successor)
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}

	err = client.Call("ChordNode.NotifySuccOfPredExit", node, &rep)
	if err != nil {
		log.Fatal("Client invocation error: ", err)
	}

	client.Close()

	close(exitChan)

	return nil
}

func (t *ChordNode) NotifySuccOfPredExit(pred *Node, reply *string) error {
	node.Predecessor = pred.Predecessor
	for k, v := range pred.Objects {
		node.Objects[k] = v
	}

	client, err := rpc.DialHTTP("tcp", node.Predecessor)
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}

	err = client.Call("ChordNode.NotifyPredOfSuccChange", node, &reply)
	if err != nil {
		log.Fatal("Client invocation error: ", err)
	}

	client.Close()

	return nil
}

func (t *ChordNode) FixSucc(arg *Arg, reply *string) error {
	node.Successor = arg.Value
	return nil
}

func (t *ChordNode) FixPred(arg *Arg, reply *string) error {
	node.Predecessor = arg.Value
	return nil
}

func main() {

	node = new(Node)
	node.Ipaddr = getMyIp()
	node.Id = mysha(node.Ipaddr)
	node.Objects = make(map[int]string)

	fmt.Println("nodo con ip " + node.Ipaddr)

	getPredAndSucc() //il nodo ottiene dal registry successore e predecessore

	enterRing() //il nodo contatta il successore e predecessore per notificarli del suo ingresso nella rete e prende la sua parte delle risorse

	node.FingerTable = make(map[int]string)
	go updateFingerTable() //il nodo aggiorna periodicamente la fingerTable (go routine)

	chordNode := new(ChordNode) //ci si mette in ascolto per ricevere un messaggio in caso di join di nodi dopo il predecessore
	rpc.Register(chordNode)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", node.Ipaddr)
	if err != nil {
		log.Fatal("Listener error: ", err)
	}
	http.Serve(listener, nil)
}
