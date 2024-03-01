# Chirpy

## What is it
A Twitter like "social network" written in Go - based on the [Learn Web Servers](https://www.boot.dev/assignments/50f37da8-72c0-4860-a7d1-17e4bda5c243) course from [boot.dev](https://www.boot.dev).

This is not a real social network, just a project to play around and get more familiar with:
- Writing http servers in Go
- Working with JSON
- Handling routing with Chi
- A little password hashing using the Crypto package
- JWT authentication
- The Go language itself

## How to compile and run
There's some env variables that need to be supplied by creating an `.env` file in the project root with a content like:
```bash
JWT_SECRET="..."
POLKA_API_KEY="..."
```


The server uses a file "database" for simplicity. It creates a database.json file in it's root directory. You can use the `--debug` flag when starting the server to enable the debug mode. Currently the only thing debug mode does is deleting the database file on startup.


To compile and start run:
```bash
go build -o out && ./out
```
