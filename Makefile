
WOLFMUD_PATH := $(patsubst %/github.com/path22/wolfmud_web, %/code.wolfmud.org/WolfMUD.git, $(PWD))

update_run:
	git pull
	go mod download
	make run

run:
	go run ./cmd/webserver/main.go

update_wolfmud:
	cd $(WOLFMUD_PATH)
	git pull
