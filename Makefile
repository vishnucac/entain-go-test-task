APP_NAME=entain_app
DB_CONTAINER=entain_db

run:
	docker-compose up -d

stop:
	docker-compose down

logs:
	docker-compose logs -f $(APP_NAME)

ps:
	docker ps

db-shell:
	docker exec -it $(DB_CONTAINER) psql -U postgres -d entain

test-init:
	curl -s http://localhost:8080/user/1/balance | jq

test-win:
	curl -X POST http://localhost:8080/user/1/transaction \
		-H "Content-Type: application/json" \
		-H "Source-Type: game" \
		-d '{"state":"win","amount":"10.15","transactionId":"tx-001"}'

test-lose:
	curl -X POST http://localhost:8080/user/1/transaction \
		-H "Content-Type: application/json" \
		-H "Source-Type: game" \
		-d '{"state":"lose","amount":"1.15","transactionId":"tx-002"}'

test-balance:
	curl -s http://localhost:8080/user/1/balance | jq
