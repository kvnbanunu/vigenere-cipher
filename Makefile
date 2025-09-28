SERVER = cmd/server/main.go
CLIENT = cmd/client/main.go
SERVER_TARGET = bin/server
CLIENT_TARGET = bin/client
BUILD = go build -o
RUN = go run
HELP = -h
MESSAGE = 'Hello, World!'
KEY = PASSWORD
IP = 127.0.0.1
PORT = 9876
SERVER_ARGS = $(IP) $(PORT)
CLIENT_ARGS = $(MESSAGE) $(KEY) $(IP) $(PORT)
COPY_CONFIG = cp config.json bin/

build-all: clean-all build-s build-c

build-s: clean-s
	@$(BUILD) $(SERVER_TARGET) $(SERVER)
	@$(COPY_CONFIG)

build-c: clean-c
	@$(BUILD) $(CLIENT_TARGET) $(CLIENT)
	@$(COPY_CONFIG)

run-s:
	@$(RUN) $(SERVER) $(SERVER_ARGS)

run-c:
	@$(RUN) $(CLIENT) $(CLIENT_ARGS)

help-s:
	@$(RUN) $(SERVER) $(HELP)

help-c:
	@$(RUN) $(CLIENT) $(HELP)

clean-s:
	@rm -f $(SERVER_TARGET)

clean-c:
	@rm -f $(CLIENT_TARGET)

clean-all:
	@rm -rf bin
