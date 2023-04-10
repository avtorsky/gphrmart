# gphrmart

[About](#about)


[![CI](https://github.com/avtorsky/gphrmart/actions/workflows/gophermart.yml/badge.svg?branch=master)](https://github.com/avtorsky/gphrmart/actions/workflows/gophermart.yml)

## About
Ecommerce loyalty program service

## Deploy

Clone repository 

```bash
git clone https://github.com/avtorsky/gphrmart.git
cd gphrmart
```

Initiate build and compile binary:

```bash
docker-compose up -d --build
cd cmd/gophermart
go build -o gophermart main.go
```

Define settings using CLI flags and init server

```bash
./gophermart --help        
Usage of ./gophermart:
  -a string
    	define service address and port (default "localhost:8090")
  -d string
    	define database connection path (default "postgres://pguser:pgpass@localhost/pgdb?sslmode=disable")
  -r string
    	define accrual system address (default "http://localhost:8080")
```

## Testing

Run unit test from root directory:

```bash
go test -v ./internal/server/handlers
```


