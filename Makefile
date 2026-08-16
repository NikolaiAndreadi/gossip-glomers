MAELSTROM_VERSION := v0.2.4
MAELSTROM_URL     := https://github.com/jepsen-io/maelstrom/releases/download/$(MAELSTROM_VERSION)/maelstrom.tar.bz2
MAELSTROM         := maelstrom/maelstrom

BIN := bin

CHALLENGES := $(notdir $(wildcard cmd/*))

# --- prerequisites

.PHONY: setup
setup:
	brew install openjdk gnuplot graphviz

$(MAELSTROM):
	curl -fsSL $(MAELSTROM_URL) | tar -xj

.PHONY: maelstrom
maelstrom: $(MAELSTROM)

# --- build

.PHONY: build
build:
	@mkdir -p $(BIN)
	@for c in $(CHALLENGES); do \
		echo ">> building $$c"; \
		go build -o $(BIN)/$$c ./cmd/$$c || exit 1; \
	done

# --- test runs

.PHONY: test-echo
test-echo: build maelstrom
	./$(MAELSTROM) test -w echo --bin $(BIN)/1-echo --node-count 1 --time-limit 10 --log-stderr

.PHONY: test-unique
test-unique: build maelstrom
	./$(MAELSTROM) test -w unique-ids --bin $(BIN)/2-unique --time-limit 30 --rate 1000 \
  --node-count 3 --availability total --nemesis partition

.PHONY: test-single-node-broadcast
test-single-node-broadcast: build maelstrom
	./$(MAELSTROM) test -w broadcast --bin $(BIN)/3a-single-node-broadcast --node-count 1 --time-limit 20 --rate 10

.PHONY: test-multi-node-broadcast
test-multi-node-broadcast: build maelstrom
	./$(MAELSTROM) test -w broadcast --bin $(BIN)/3b-multi-node-broadcast --node-count 5 --time-limit 20 --rate 10

.PHONY: test-efficient-fault-tolerant-broadcast
test-efficient-fault-tolerant-broadcast: build maelstrom
	./$(MAELSTROM) test -w broadcast --bin $(BIN)/3e-efficient-broadcast --node-count 25 --time-limit 20 --rate 100 \
  --latency 100 --nemesis partition

.PHONY: test-grow-only-counter
test-grow-only-counter: build maelstrom
	./$(MAELSTROM) test -w g-counter --bin $(BIN)/4-grow-only-counter --node-count 3 --rate 100 --time-limit 20 \
  --nemesis partition

.PHONY: test-grow-only-counter-stateless
test-grow-only-counter-stateless: build maelstrom
	./$(MAELSTROM) test -w g-counter --bin $(BIN)/4-grow-only-counter-stateless --node-count 3 --rate 100 --time-limit 20 \
  --nemesis partition

.PHONY: test-single-node-log
test-single-node-log: build maelstrom
	./$(MAELSTROM) test -w kafka --bin $(BIN)/5a-single-node-kafka-style-log --node-count 1 --concurrency 2n \
  --time-limit 20 --rate 1000

.PHONY: test-distributed-log
test-distributed-log: build maelstrom
	./$(MAELSTROM) test -w kafka --bin $(BIN)/5b-distributed-kafka-style-log --node-count 2 --concurrency 2n \
  --time-limit 20 --rate 1000

.PHONY: test-efficient-log
test-efficient-log: build maelstrom
	./$(MAELSTROM) test -w kafka --bin $(BIN)/5c-efficient-kafka-style-log --node-count 2 --concurrency 2n \
  --time-limit 20 --rate 1000

# --- results

.PHONY: serve
serve: maelstrom
	./$(MAELSTROM) serve

.PHONY: clean
clean:
	rm -rf $(BIN) store
