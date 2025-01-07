# Shade Web Server

##  Solution Design

<image src="./docs/clearn-architecture.png"  alt="clean architecture diagram"/>

<br>

The used design is a minified version of "Clean Architecture", it will make sence when we map each architecture layer with the directory structure.
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

> **The Most Important note is** data flows only in one direction, this means that inner layers cannot call functions from the outer layers, the only exception is the applicaiton layer can call the presistence layer and it is done through interface not a direct call, that is the only exception, and here is and example of data flow:

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

Everything things is configured in the docker-compose file including the live reload to watch for our development updated and the database configuration, so it is as simple as running:

`docker-compose up --build -d`

If it is the initial run you should also run the following migrations command to update the schema of the database

`docker exec -it app go run main.go migrate up`

Then you can visit the server using that url: "http://localhost:8080"