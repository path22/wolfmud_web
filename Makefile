
WOLFMUD_PATH := $(patsubst %/github.com/path22/wolfmud_web, %/code.wolfmud.org/WolfMUD.git, $(PWD))

run:
	go run ./cmd/webserver/main.go

update:
	git pull
	go mod download

update_wolfmud:
	cd $(WOLFMUD_PATH)
	git pull
