project ?= echo
maelstrom = .maelstrom/maelstrom

.PHONY: build run-echo run-unique-ids serve

build:
	@echo "Building $(project)"
	@cd $(project) && go build
	

# run echo server with a single maelstrom node for 10 seconds
run-echo:
	$(maelstrom) test -w echo --bin ./echo/echo --node-count 1 --time-limit 10


# run a 3-node highly available cluster for 3 seconds and injecting request payloads at a rate of 10k per second while inducing network partitions.
# all ids will be verified as globally unique
run-unique-ids:
	$(maelstrom) test -w unique-ids --bin ./unique-ids/unique-ids --time-limit 30 --rate 10000 --node-count 3 --availability total --nemesis partition


# view maelstrom results
serve:
	$(maelstrom) serve