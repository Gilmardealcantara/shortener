.PHONY: test
test:
	go test -v ./...

.PHONY: run
run:
	go run main.go

.PHONY: migrateup
migrateup:
	migrate -path db/migrations -database "postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable" up

.PHONY: migratedown
migratedown:
	migrate -path db/migrations -database "postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable" down

.PHONY: pgcon
pgcon:
	docker exec -it local_postgres psql -U myuser -d mydb

.PHONY: rediscon
rediscon:
	docker exec -it local_redis redis-cli

.PHONY: curl_post
curl_post:
	curl -X POST http://localhost:8080/shorten -H "Content-Type: application/json" -d '{"long_url": "http://pudim.com"}'

.PHONY: curl_get
curl_get: 
	curl -v -X GET http://localhost:8080/100M
