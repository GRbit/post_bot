PID      = /tmp/holodok-secret.pid
GO_FILES = $(find -iname '*.go' )
PKG      = .
BIN      = ./bot
ENV      = ./.env

fmt:
	gofumpt -l -w -extra ./.
	gci write -s Std -s 'Prefix(github.com/grbit/post_bot)' -s Def .

serve: restart
	@fswatch -x -o --event Created --event Updated --event Renamed -r -e '.*' -i '\.go' . | \
		xargs -n1 -I{}  make restart || make stop

serve-docker: restart
	fswatch -x -or --event Created --event Updated --event Renamed . | \
		xargs -n1 -I{}  make restart || make stop

start:
	@$(BIN) & echo $$! > $(PID)

stop:
	@{ kill -9 `cat $(PID)` && wait `cat $(PID)`; } 2> /dev/null || true

build: $(GO_FILES)
	@printf "Building..."
	@go build -o $(BIN) $(PKG)
	@echo " done"

restart: stop build start

test:
	@GLUE_ENV=test go test ./... -race

.PHONY: serve serve-docker stop start restart test

docker-db-create:
	@docker run --name postmeta-pgsql --env-file $(ENV) -p 5432:5432 -d postgres

pre-push-hook:
	@./githooks/pre-push.sh

install-pre-push:
	@grep 'make pre-push-hook' .git/hooks/pre-push 2> /dev/null; \
	if [ $$? = 0 ]; then \
		echo "already installed"; \
	elif [ -f ".git/hooks/pre-push" ]; then \
		echo ".git/hooks/pre-push exists. Consider it to be a text bash file."; \
		echo "make pre-push-hook" >> .git/hooks/pre-push; \
	else \
		echo "#!/bin/sh\n\nmake pre-push-hook" > .git/hooks/pre-push; \
	fi; \
	chmod +x ".git/hooks/pre-push"
