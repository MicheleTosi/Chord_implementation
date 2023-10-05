@echo off
setlocal

REM Chiedi all'utente l'indirizzo dell'istanza EC2
set /p ec2_address="Inserisci l'indirizzo IP pubblico dell'istanza EC2: "

REM Chiedi all'utente il percorso alla chiave SSH
set /p ssh_key="Inserisci il percorso alla chiave SSH (ad esempio, 'sdcc_key.pem'): "

REM Chiedi all'utente il percorso del progetto da caricare sull'istanza EC2
set /p proj="Inserisci il percorso al progetto da caricare sull'istanza EC2: "

REM Carica la cartella sulla tua istanza EC2
echo Caricamento dei file su EC2...
echo "yes\n" | scp -i %ssh_key% -r %proj% ec2-user@%ec2_address%:/home/ec2-user/

REM Esegui il comando SCP in un nuovo terminale (se hai il terminale adatto)
start cmd /k "ssh -i %ssh_key% ec2-user@%ec2_address%"

REM Connessione SSH all'istanza EC2
echo Connessione a EC2...
ssh -i %ssh_key% ec2-user@%ec2_address%

