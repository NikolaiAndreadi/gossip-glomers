MAELSTROM_VERSION := v0.2.4
MAELSTROM_URL     := https://github.com/jepsen-io/maelstrom/releases/download/$(MAELSTROM_VERSION)/maelstrom.tar.bz2
MAELSTROM_TARBALL := maelstrom.tar.bz2
MAELSTROM         := ./maelstrom/maelstrom
MTEST             := $(MAELSTROM) test

BIN := bin

GO_FILES := $(shell find cmd -name '*.go') go.mod go.sum

.DEFAULT_GOAL := build
.DELETE_ON_ERROR:

# --- prerequisites

.PHONY: setup
setup:
	brew install openjdk gnuplot graphviz

$(MAELSTROM):
	@rm -rf maelstrom $(MAELSTROM_TARBALL)
	curl -fsSL $(MAELSTROM_URL) -o $(MAELSTROM_TARBALL)
	tar -xjf $(MAELSTROM_TARBALL)
	@rm -f $(MAELSTROM_TARBALL)

.PHONY: maelstrom
maelstrom: $(MAELSTROM)

# --- build

.PHONY: build
build:
	go build -o $(BIN)/ ./cmd/...

$(BIN)/%: $(GO_FILES)
	go build -o $@ ./cmd/$*

# --- test runs

.PHONY: test-echo
test-echo: $(BIN)/1-echo $(MAELSTROM)
	$(MTEST) -w echo --bin $< --node-count 1 --time-limit 10 --log-stderr

.PHONY: test-unique
test-unique: $(BIN)/2-unique $(MAELSTROM)
	$(MTEST) -w unique-ids --bin $< --time-limit 30 --rate 1000 \
  --node-count 3 --availability total --nemesis partition

.PHONY: test-single-node-broadcast
test-single-node-broadcast: $(BIN)/3a-single-node-broadcast $(MAELSTROM)
	$(MTEST) -w broadcast --bin $< --node-count 1 --time-limit 20 --rate 10

.PHONY: test-multi-node-broadcast
test-multi-node-broadcast: $(BIN)/3b-multi-node-broadcast $(MAELSTROM)
	$(MTEST) -w broadcast --bin $< --node-count 5 --time-limit 20 --rate 10

.PHONY: test-efficient-fault-tolerant-broadcast
test-efficient-fault-tolerant-broadcast: $(BIN)/3e-efficient-broadcast $(MAELSTROM)
	$(MTEST) -w broadcast --bin $< --node-count 25 --time-limit 20 --rate 100 \
  --latency 100 --nemesis partition

.PHONY: test-grow-only-counter
test-grow-only-counter: $(BIN)/4-grow-only-counter $(MAELSTROM)
	$(MTEST) -w g-counter --bin $< --node-count 3 --rate 100 --time-limit 20 \
  --nemesis partition

.PHONY: test-grow-only-counter-stateless
test-grow-only-counter-stateless: $(BIN)/4-grow-only-counter-stateless $(MAELSTROM)
	$(MTEST) -w g-counter --bin $< --node-count 3 --rate 100 --time-limit 20 \
  --nemesis partition

.PHONY: test-single-node-log
test-single-node-log: $(BIN)/5a-single-node-kafka-style-log $(MAELSTROM)
	$(MTEST) -w kafka --bin $< --node-count 1 --concurrency 2n \
  --time-limit 20 --rate 1000

.PHONY: test-distributed-log
test-distributed-log: $(BIN)/5b-distributed-kafka-style-log $(MAELSTROM)
	$(MTEST) -w kafka --bin $< --node-count 2 --concurrency 2n \
  --time-limit 20 --rate 1000

.PHONY: test-efficient-log
test-efficient-log: $(BIN)/5c-efficient-kafka-style-log $(MAELSTROM)
	$(MTEST) -w kafka --bin $< --node-count 2 --concurrency 2n \
  --time-limit 20 --rate 1000

# --- results

.PHONY: serve
serve: $(MAELSTROM)
	$(MAELSTROM) serve

.PHONY: clean
clean:
	rm -rf $(BIN) store $(MAELSTROM_TARBALL)
