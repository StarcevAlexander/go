для запуска на сервере

apt update && apt upgrade -y
apt install -y wget curl nano htop
apt install -y git
git --version 
git clone https://github.com/StarcevAlexander/go.git

wget https://golang.org/dl/go1.21.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version

go build


go mod init github.com/StarcevAlexander/go

nohup ./myapp > server.log 2>&1 & 
//запускает в фоне
