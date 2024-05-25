# user-auth

**User Authentication with Options**

This repository provides user authentication functionalities with a choice of MongoDB or SQLite for data storage.

* **MongoDB Authentication:**
  * Branch: [mongodb](https://github.com/BryanSouza91/user-auth/tree/mongodb) (or specific branch name)
  * User authentication and session tokenization using Redis and MongoDB.
* **SQLite Authentication:**
  * Branch: [sqlite](https://github.com/BryanSouza91/user-auth/tree/sqlite) (or specific branch name)
  * User authentication and session tokenization using Redis and SQLite.

**Choosing the Database:**

Select the branch that aligns with your preferred database system. 

**Additional Notes**

Work in progress 

## Instructions

### Install Go Programming language latest version
[Download and install here](https://go.dev/)

[![N|Solid](https://sdtimes.com/wp-content/uploads/2018/02/golang.sh_-490x490.png)](https://golang.org/dl/)

### External Packages
* [redis](https://github.com/gomodule/redigo/redis) - Redis Server
* [mongo-driver](https://pkg.go.dev/go.mongodb.org/mongo-driver) - MongoDB driver
* [go-sqlite3](https://pkg.go.dev/mattn/go-sqlite3) - SQLite3 driver

### Run Redis Server
Before running, a Redis Server instance is required to be running.
