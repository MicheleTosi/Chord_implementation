# Chord_implementation
Progetto finale per il corso di Sistemi Distribuiti e Cloud Computing, facoltà di Ingegneria Informatica Magistrale, Università di Roma Tor Vergata.
<br>Il sistema prevede la realizzazione di una overlay network strutturata gestita tramite il protocollo Chord.
</br>Componenti del sistema:
<ul>
<li>Nodi: si preoccupano della gestione delle risorse (rimozione, inserimento, ricerca).
<li>Client: si interfaccia con i noid per eseguire le operazioni sul sistema, si preoccupa inoltre dell'eventuale rimozione dei nodi.
<li>Server registry: si preoccupa di inviare al client l'indirizzo IP di un nodo della rete oppure di inviare l'indirizzo IP dei vicini al nodo che vuole entrare nella rete. Gestisce inoltre il caso in cui un nodo fallisca.
</ul>

# Esecuzione del progetto
## Inizializzazione del sistema
Per eseguire il programma è necessario avviare il Docker Server.<br>
In seguito utilizzare il file `config.json` per definire m (numero di bits utilizzati per ID di nodi e risorse) e nodes (il numero di nodi che vengono creati all'avvio).<br>
Una volta fatto questo aggiornare il file `docker-compose.yml` attraverso il comando 
``` 
go run generate_compose.go
```
A questo punto eseguire il build 
```
docker-compose build
```
e avviare 
```
docker-compose.up
```
Dopo questi comandi il sistema sarà pronto per essere utillizzato.
## Client
Per eseguire il client eseguire il comando 
```
cd client
go run client.go
```
## Creazione nuovo nodo
Per creare un nuovo nodo dopo che il programma è in esecuzione basta eseguire il comando
```
./start_node.sh [x]
```
dove x è un intero utilizzato per calcolare la porta del nuovo nodo che entra nel sistema.
# EC2
Per eseguire il programma su EC2 si deve:
- Creare una nuova istanza EC2 direttamente dal sito di AWS.
- Aprire il prompt dei comandi da Windows ed eseguire il comando inserendo come richiesto IP pubblico dell'istanza EC2, il percorso alla chiave SSH e quello alla cartella contenente il progetto (quando si inserisce il collegamento utilizzare '/' al posto di '\').
```
initializa_aws.bat
```
- Eseguire su una scheda del terminale il comando seguente in modo da preparare l'ambiente di esecuzione
```
sudo chmod +x configure_aws.sh
```
- Sull'altra scheda di terminale aperta avviare il client.

Per pulire l'ambiente dei container utilizzare il comando 
```
docker container prune
```
Scollegarsi dall'istanza con il comando 
```
exit
```