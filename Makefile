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
COPY_TESTING = cp testing/Makefile bin/

all: clean buildserver buildclient
	@$(COPY_TESTING)

buildserver:
	@$(BUILD) $(SERVER_TARGET) $(SERVER)
	@$(COPY_CONFIG)

buildclient:
	@$(BUILD) $(CLIENT_TARGET) $(CLIENT)
	@$(COPY_CONFIG)

runserver:
	@$(RUN) $(SERVER) $(SERVER_ARGS)

runclient:
	@$(RUN) $(CLIENT) $(CLIENT_ARGS)

clean:
	@rm -rf bin
