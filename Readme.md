
migrate create -ext sql -dir db/migrations -seq init_schema

 docker exec -it local_postgres psql -U myuser -d mydb
