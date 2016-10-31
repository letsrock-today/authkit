GOLDFLAGS ?= "-w -s"

.PHONY: \
	all \
	clean \
	test \
	build \
	up \
	down

all: \
	clean \
	test \
	build
	@echo "All Done."

clean:
	@echo "Clean"
	@rm -rf ./.build
	@go clean -i -r \
		github.com/letsrock-today/authkit/sample/authkit/backend
	@echo "Clean Done."

test:
	@echo "Test"
	@glide nv | xargs go test
	@-glide nv | xargs go vet
	@-glide nv | xargs golint
	@echo "Test Done."

build:
	@echo "Build"
	@rm -rf ./.build
	@mkdir -p ./.build
	@go build -ldflags $(GOLDFLAGS) \
		-o ./.build/backend/authkit-sample \
		github.com/letsrock-today/authkit/sample/authkit/backend
	@cd ./ui-web && npm run dist -- --output-path ../.build/ui-web/dist
	@cp -r ./ui-web/html ./.build/ui-web
	@cp -r ./ui-web/templates ./.build/ui-web
	@cp -r ./backend/env ./.build/backend
	@echo "Build Done."

up: clean \
	build 
	@echo "Starting App"
	@cd ./docker-compose && sudo docker-compose up -d
	@cd ./.build/backend && ./authkit-sample $(ARGS) >app.log 2>app.err.log &
	@sleep 5
	-sensible-browser "https://localhost:8080" >/dev/null 2>/dev/null &
	-sensible-browser "https://localhost:8080/oauth2/auth?client_id=a8a9bffa-d748-440d-9d67-bd161018a495&response_type=code&scope=core+hydra.keys.get&state=vboeidlizlxrywkwlsgeggff" >/dev/null 2>/dev/null &
	@echo "App is running."

down:
	@echo "Stopping App"
	@-pkill authkit-sample
	@cd ./docker-compose && sudo docker-compose down -v --rmi local
	@echo "App is stopped."
