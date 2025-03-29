# gopostgresql Example

A simple TODO list based on Emmanuel John blog post: https://blog.logrocket.com/building-simple-app-go-postgresql/

gopostgresql uses gofiber for the web framework.

gopostgresql uses pgAdmin to manage our Postgres database visually.

pgAdmin4:
- brew update
- brew install pgadmin4
- database: postgres
- password: DaVinci
- create database: todolist (owner: postgres, password: DaVinci)
- port: 5432

Test:

    psql -U postgres -p 5432 -h localhost todos

Create:
- database: todos
- table: todos
- column: item <text>

    mkdir ~/pixelmonks/go/gopostgresql
    cd ~/pixelmonks/go/gopostgresql
    go mod init gopostgresql
    go mod tidy

## Air: used for hot reload:

    go get github.com/air-verse/air

(Required update from: go get github.com/cosmtrek/air)

## Run:

    go run github.com/air-verse/air

(Required update from: go run github.com/cosmtrek/air)

## Web:

http://localhost:3000

or

http://127.0.0.1:3000
