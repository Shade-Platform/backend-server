# Shade Web Server

##  Solution Design

<image src="./docs/clearn-architecture.png"  alt="clean architecture diagram"/>

<br>

The used design is a minified version of "Clean Architecture", it will make sense when we map each architecture layer with the directory structure.
- Presentation layer: routers directory
- Persistence, Application, Domain layers: all lives in core directory

Here is a list of the role of each layer:
- Presentation Layer: the layer that faces the outside world that provides API endpoints or GUI, but of course in this server solution it provides APIs only.
- Application Layer: Implements business logic.
- Persistence Layer: Implements functions to store data in the database.
- Domain Layer: Defines the the sturcture of each model.

Notes:
- These layers represents the flow of data when the API endpoint is hit till how it is processed.
- As mentioned before, this is a simplified version of the clean architecture as this solution is developed for demo puposes.

> **The Most Important note is** data flows only in one direction, this means that inner layers cannot call functions from the outer layers, the only exception is the applicaiton layer can call the presistence layer and it is done through interface not a direct call, that is the only exception, and here is an example of data flows:

<image src="./docs/dfd.png"  alt="data flow diagram"/>

### Directory Structure

> Note: user implementation files are created as example and to be a reference.

```
project/
│
├── core/                           
│   └── users/                      # domain, application, and presistence layers layers 
│       ├── user.go                 # model main attributes
│       ├── user_service.go         # business logic functions
│       └── user_repository.go      # communication with the database
│
├── routers/                        # presentation layer, depends on core logic
│   └── users_route.go                # apis
│
├── infrastructure/                 # infrastructure configuraiton
│   ├── migrations/                 # Migration scripts
│   │   └── 20230101010101_create_users_table.up.sql
│   └── db_connection.go
│
├── main.go                         # Entry point that wires everything together
│
├── go.mod                          # Go module definition
└── go.sum                          # Dependency tracking
```

## How to Run

**Prerequests:** Docker, Docker Compose

Everything things is configured in the docker-compose file including the live reload to watch our development updates and the database configuration, so it is as simple as running:

`docker-compose up --build -d`

If it is the initial run you should also run the following migrations command to update the schema of the database. This command will run all migrations - contact danny if you need more help on the different migration commands

`docker exec -it app go run main.go migrate up`

Please make sure that you have the minikube cluster running beforehand, otherwise the cluster connection won't work and return an error. The cluster configuration is located at .kube/config.local, which tries to access ~/.minikube, make sure that .minikube exists. (The docker file tries to link them to the container beforehand)

Then you can visit the server using that url: "http://localhost:8080" (Of course there is nothing on the route "/", to check if the server is running: http://localhost:8080/health)

If there's an erorr with migrations:
- instead of running migrations, just dump data manually using this command

`docker exec -i mysql_db mysqldump -u root -ppassword mydb | tee ./mydb_dump.sql`



- to run the database: `docker-compose up --build`
- to run go server: `go run main.go -kuberconfig ~/.kube/config`
