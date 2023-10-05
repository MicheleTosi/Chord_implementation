package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sort"
	"time"
)

//il registry riceve richieste dal client per quanto riguarda il nodo da contattare per inserire, cercare, rimuovere risorse
//l'ip verrà ritornato con Round Robin per evitare per esempio uno scheduling random che potrebbe appesantire troppo solo alcuni nodi

//il registry riceve richieste dai nodi che entrano nella rete per trovare il successore

type Neighbors struct {
	Predecessor string
	Successor   string
}

type Arg struct {
	Id    int
	Value string
}

type Registry string

var Nodes = make(map[int]string)

//il registry riceve richieste dal client per quanto riguarda il nodo da contattare per inserire, cercare, rimuovere risorse
//l'ip verrà ritornato con Round Robin per evitare per esempio uno scheduling random che potrebbe appesantire troppo solo alcuni nodi

/*Neighbors
* attraverso questa funzione i nodi ricevono dal registry ip del predecessore e del successore
 */
func (t *Registry) Neighbors(arg *Arg, reply *Neighbors) error {
	m, err := ReadFromConfig()
	if err != nil {
		log.Fatal("Errore nella lettura dal file config.json: ", err)
	}
	maxNodi := 1 << m
	id := arg.Id
	if len(Nodes) == 0 { //controllo se è il primo nodo ad essere inserito, in questo caso predecessore e successore coincidono con l'ip del nodo appena aggiunto
		Nodes[id] = arg.Value
		reply.Successor = arg.Value
		reply.Predecessor = arg.Value
		fmt.Println(Nodes)
		return nil
	}
	if Nodes[id] != "" { //controllo se non esiste già un nodo con lo stesso ID
		reply = nil
		return errors.New("Esiste già un nodo con questo ID")
	}
	if len(Nodes) >= maxNodi { //controllo se non è già stato raggiunto il numero massimo di nodi
		reply = nil
		return errors.New("È stato raggiunto il numero massimo di nodi")
	}

	keys := make([]int, 0, len(Nodes)) //mi creo una slice delle chiavi della mappa
	for k := range Nodes {
		keys = append(keys, k)
	}

	//ordino la slice contenente le chiavi
	sort.Sort(sort.IntSlice(keys))

	//inserisco il nodo nella map
	Nodes[id] = arg.Value
	fmt.Println(Nodes)

	//cerco l'ip del successore e del predecessore lo ritorno
	for i, k := range keys {
		if id < k {
			reply.Successor = Nodes[k]
			if i-1 >= 0 {
				reply.Predecessor = Nodes[keys[i-1]]
			} else {
				reply.Predecessor = Nodes[keys[len(keys)-1]]
			}
			return nil
		}
	}
	reply.Successor = Nodes[keys[0]]             //se l'id del nodo che sto inserendo è il maggiore ritorno l'ip del primo nodo dell'anello
	reply.Predecessor = Nodes[keys[len(keys)-1]] //se l'id del nodo che sto inserendo è il maggiore ritorno l'ip dell'ultimo nodo dell'anello

	return nil
}

/*ReturnChordNode
*funzione invocata dal client per ottenere un nodo della rete
*la politica seguita per la scelta del nodo è Round Robin in modo da cercare di mantenere il carico fair
 */

var lastNodeSelected = -1

func (t *Registry) ReturnChordNode(arg *Arg, reply *string) error {
	if len(Nodes) == 0 {
		return errors.New("Non ci sono nodi nella rete")
	}
	keys := make([]int, 0, len(Nodes))
	for k := range Nodes {
		keys = append(keys, k)
	}
	lastNodeSelected = (lastNodeSelected + 1) % len(keys)
	*reply = Nodes[keys[lastNodeSelected]]
	return nil
}

/*GetNodes
*funzione utilizzata dal client per conoscere quali sono i nodi presenti nella rete e decidere in seguito quale eliminare
 */
func (t *Registry) GetNodes(arg *Arg, reply *string) error {
	if len(Nodes) == 0 {
		*reply = "Non sono presenti nodi nella rete."
		return nil
	}
	*reply = "Nodi presenti:"
	for k, v := range Nodes {
		*reply = fmt.Sprintf("%s\n%d: %v", *reply, k, v)
	}
	return nil
}

/*RemoveNode
*funzione invocata dal client per rimuovere un nodo dalla rete
 */
func (t *Registry) RemoveNode(arg *Arg, reply *string) error {
	if len(Nodes) == 0 {
		*reply = "Non ci sono nodi nella rete."
		return nil
	}
	if len(Nodes) == 1 {
		*reply = "Non è possibile rimuovere il nodo in quanto è l'unico presente nella rete."
		return nil
	}

	nodeId := arg.Id
	if Nodes[nodeId] == "" {
		*reply = "Non è presente nella rete nessun nodo con l'id specificato."
		return nil
	}

	client, err := rpc.DialHTTP("tcp", Nodes[nodeId]) //contatto il nodo da rimuovere in modo che lui possa aggiornare i vicini sul fatto che verrà rimosso
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}

	err = client.Call("ChordNode.ExitRing", arg, &reply)
	if err != nil {
		log.Fatal("Client invocation error: ", err)
	}

	client.Close()

	delete(Nodes, nodeId)
	fmt.Println(Nodes)
	*reply = fmt.Sprintf("Nodo con id '%d' rimosso correttamente", nodeId)

	return nil
}

func (t *Registry) IsNodeAlive(arg *Arg, reply *int) error {
	node := arg.Value
	id := arg.Id
	if Nodes[id] != "" {
		client, err := net.DialTimeout("tcp", node, 5*time.Second)
		if err != nil {
			fmt.Printf("Non riesco a contattare [%d:%s], procedo con la sua rimozione.\n", id, node)
			FixNeighbors(id)
			delete(Nodes, id)
			fmt.Println(Nodes)
		} else {
			client.Close()
		}

	}
	return nil
}

func FixNeighbors(id int) {
	var succ string
	var pred string
	var reply string
	arg := new(Arg)
	keys := make([]int, 0, len(Nodes))
	for k := range Nodes {
		keys = append(keys, k)
	}
	sort.Sort(sort.IntSlice(keys))

	for i, v := range keys {
		if v == id {
			if i == 0 {
				pred = Nodes[keys[len(keys)-1]]
				succ = Nodes[keys[1]]
			} else if i == len(keys)-1 {
				pred = Nodes[keys[len(keys)-2]]
				succ = Nodes[keys[0]]
			} else {
				pred = Nodes[keys[i-1]]
				succ = Nodes[keys[i+1]]
			}
		}
	}

	arg.Value = succ
	client, err := rpc.DialHTTP("tcp", pred) //contatto il nodo da rimuovere in modo che lui possa aggiornare i vicini sul fatto che verrà rimosso
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}

	err = client.Call("ChordNode.FixSucc", arg, &reply)
	if err != nil {
		log.Fatal("Client invocation error: ", err)
	}

	client.Close()

	arg.Value = pred
	client, err = rpc.DialHTTP("tcp", succ) //contatto il nodo da rimuovere in modo che lui possa aggiornare i vicini sul fatto che verrà rimosso
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}

	err = client.Call("ChordNode.FixPred", arg, &reply)
	if err != nil {
		log.Fatal("Client invocation error: ", err)
	}

	client.Close()
}

func main() {
	registry := new(Registry)
	rpc.Register(registry)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", "registry:1234")
	if err != nil {
		log.Fatal("Listener error: ", err)
	}
	http.Serve(listener, nil)
}
