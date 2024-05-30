.PHONY: build



build:
	go build -o ./build/itashi ./cmd/itashi/main.go 

install:
	sudo mv ./build/itashi /usr/local/bin/itashi && sudo chmod u+x /usr/local/bin/itashi

uninstall:
	sudo rm -f /usr/local/bin/itashi
