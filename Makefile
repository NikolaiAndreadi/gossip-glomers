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
	./$(MAELSTROM) test -w unique-ids --bin $(BIN)/3a-single-node-broadcast --time-limit 30 \
  --rate 1000 --node-count 3 --availability total --nemesis partition
# --- results

.PHONY: serve
serve: maelstrom
	./$(MAELSTROM) serve

.PHONY: clean
clean:
	rm -rf $(BIN) store
