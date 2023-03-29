build:
	@cd cmd/serve && go build

run: export ADDRESS = :8888
run: export CAPACITY = 8
run: export GIN_MODE = release
run: export TIMEOUT = 3s
run:
	@cd cmd/serve && ./serve

testing:
	curl http:/localhost:8888/v1/greet && echo
	curl -X PUT -H "Content-Type: application/json" -d '{"Greeting":"hi"}' http:/localhost:8888/v1/greeting && echo
	curl http:/localhost:8888/v1/greet && echo

