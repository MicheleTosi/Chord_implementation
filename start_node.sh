#!/bin/bash

x=${1:-0}  # Se non viene passato nessun valore x viene messo a 0

# Legge dal file config.json il numero di nodi da creare
m=$(cat config.json | jq -r '.nodes')

# Calcola la porta del nuovo nodo facendo '8000+m+x'
node_number=$((m + x +1 ))
port=$((8000 + node_number))

# Fa il build della docker immagine
sudo docker build -t "sdcc_chord_node${node_number}" -f DockerFiles/node/Dockerfile .

# Esegue il container docker con la porta calcolata (--rm permette il riavvio del container senza usare i flag).
sudo docker run --rm -p $port:$port --network=sdcc_chord_my_network -e NODE_PORT=$port "sdcc_chord-node${node_number}"