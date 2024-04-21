.DEFAULT_GOAL := build
.MAIN :-= build

build:
	cd cmd/mygit && go build -o ../../bin/mygit

clean:
	rm -rf bin/