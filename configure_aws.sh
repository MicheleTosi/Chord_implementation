# Il resto degli script rimane invariato
sudo yum update -y
sudo yum install docker -y
sudo yum install go -y
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
sudo systemctl start docker
sudo systemctl enable docker
sudo docker-compose build
sudo docker-compose up