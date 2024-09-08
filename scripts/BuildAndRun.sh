# this is mostly a test script
# setups everything for server, builds it and ther runs the server. 

sudo docker pull postgres
sudo docker run --name=goproj-db -e POSTGRES_PASSWORD='qwerty' -p 5436:5432 -d --rm postgres
sudo docker ps

dbmate up

go build ./cmd/main.go
./main
