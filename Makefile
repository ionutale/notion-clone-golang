.PHONY: dev build web-build web-dev run clean

dev: web-dev
	@echo "Run Go server in another terminal: go run ."

web-dev:
	cd web && pnpm dev &

web-build:
	cd web && pnpm build

build: web-build
	go build -o bin/notion-clone .

run: build
	./bin/notion-clone

clean:
	rm -rf bin/ web/build/ web/node_modules/
