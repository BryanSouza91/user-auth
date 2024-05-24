# user-auth

Work in progress 

User authentication and session tokenization using Redis and SQLite.

# Instructions

### Install Go Programming language latest version
[Download and install here](https://go.dev/)

[![N|Solid](https://sdtimes.com/wp-content/uploads/2018/02/golang.sh_-490x490.png)](https://golang.org/dl/)

### External Packages
* [redis](https://github.com/gomodule/redigo/redis) - Redis Server
* [go-sqlite3](https://pkg.go.dev/mattn/go-sqlite3) - SQLite3 driver

### Run Redis Server
Before running, a Redis Server instance is required to be running.

### Configuration .json
Besure to recreate your own conf.json using the example provided

### To get this repository and run

 ```sh
$ git clone https://github.com/BryanSouza91/user-auth.git
$ go run app.go
```
