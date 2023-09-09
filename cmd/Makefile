postgresinit:
	docker run --name matchme_db -e POSTGRES_PASSWORD=qwerty -p 5436:5432 -d postgres 

postgresrm:
	docker rm -f matchme_db

.PHONY: postgresinit, postgresrm