#Back-end for a simple e-shop

Note: This repository is a part of the Computer Science Capstone project made by the [University of Phoenix](https://www.phoenix.edu/) student [Elena Lazar](https://www.linkedin.com/in/helen-lazar-36315a95/).

This is a simple monolith back-end for a simple e-commerce solution providing a REST API for a front-end application. The service stores and retrieves data from a database, and secures some of its endpoint using authentication (at the initial stage, only [OAUth2 provided by Google](https://console.cloud.google.com/apis/credentials)). To speed up the development, the boilerplate code is generated by [go-swagger](https://goswagger.io/) tool, and [Bun ORM](https://bun.uptrace.dev/) is used for database interactions.

The front-end code base for this solution is located at [https://bitbucket.org/morrah/estore-frontend](https://bitbucket.org/morrah/estore-frontend)

Endpoints:
  - Authentication:
    - login
    - OAUth2 redirect
  - Products:
    - list (pageable, searchable)
    - get by ID
    - add (secured by admin scope)
    - update (secured by admin scope)
    - delete (secured by admin scope)
  - Categories:
    - list (pageable, searchable)
    - get by ID
    - add (secured by admin scope)
    - update (secured by admin scope)
    - delete (secured by admin scope)
  - Orders:
    - list (pageable, secured by private/admin scopes)
    - get by ID (secured by private/admin scopes)
    - add (secured by private/admin scopes)
    - update (secured by private/admin scopes)
    - delete (secured by private/admin scopes)
  - Users
    - list (pageable, searchable, secured by admin scope)
    - get by ID (secured by private/admin scopes)
    - add (secured by private/admin scopes)
    - update (secured by private/admin scopes)
    - delete (secured by private/admin scopes)
  - Payments:
    - list (pageable, secured by private/admin scopes)
    - get by ID (secured by private/admin scopes)
    - add (secured by private/admin scopes)
    - update (secured by private/admin scopes)
    - delete (secured by private/admin scopes)

##Development

Validate the OpenAPI specification before generating the server code

```
swagger validate ./swagger.yml
```

Generate server boilerplate code

```
swagger generate server -t server -A EStoreMain -P models.Principal -f ./swagger.yml
```

Update your dependencies if needed

```
go mod tidy
```

Implement your changes

Set up your OAuth2 credentials

Fill your config.json file with database type and connection string, as well as with your OAuth2 credentials

Run the service locally:

```
go run ./server/cmd/e-store-main-server/main.go --port 8080 --tls-certificate ./certs/server.crt --tls-key ./certs/server.key --tls-port 8443
```

#Build

mkdir bin

##For Linux (production)

go build -o ./bin/server ./server/cmd/e-store-main-server/main.go

###Run 
./bin/server --port 8080 --tls-certificate ./certs/server.crt --tls-key ./certs/server.key --tls-port 8443


##For Windows (Lab environment)

env GOOS=windows GOARCH=amd64 go build -o ./bin/server.exe ./server/cmd/e-store-main-server/main.go

###Run

bin\server.exe --port 8080 --tls-certificate ./certs/server.crt --tls-key ./certs/server.key --tls-port 8443
