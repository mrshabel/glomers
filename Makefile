project ?= echo
maelstrom = .maelstrom/maelstrom

.PHONY: build run-echo serve

build:
	@echo "Building $(project)"
	@cd $(project) && go build
	

# run echo server with a single maelstrom node for 10 seconds
run-echo:
	$(maelstrom) test -w echo --bin ./echo/echo --node-count 1 --time-limit 10

# view maelstrom results
serve:
	$(maelstrom) serve