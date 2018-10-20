GOCMD   =  go
GOBUILD =  $(GOCMD)  build
GOINSTALL =  $(GOCMD)  install
GOCLEAN =  $(GOCMD)    clean
GOTEST  =  $(GOCMD)    test
GOGET   =  $(GOCMD)    get
GOSET   =              set
GOFILES	=	$(wildcard *.go)
CLINET_PATH =  ./cmd/client
CLINET_BINARY_NAME_WIN =  client.exe
CLIENT_BINARY_NAME_LIN =  client_lin_out
SERVER_PATH =  ./cmd/server
SERVER_BINARY_NAME_WIN =  server.exe
SERVER_BINARY_NAME_LIN =  server_lin_out


client-build-win:
	set GOOS=windows
	set GOARCH=amd64
	$(GOBUILD)	-o $(CLINET_PATH)/$(CLINET_BINARY_NAME_WIN) -v $(CLINET_PATH)/$(GOFILES)
client-build-lin:
	set GOOS=linux
	set GOARCH=amd64
	$(GOBUILD)	-o $(CLINET_PATH)/$(CLIENT_BINARY_NAME_LIN) -v $(CLINET_PATH)/$(GOFILES)
client-clean:
	$(GOCLEAN)	$(CLINET_PATH)
server-build-win:
	set GOOS=windows
	set GOARCH=amd64
	$(GOBUILD)	-o $(SERVER_PATH)/$(SERVER_BINARY_NAME_WIN) -v $(SERVER_PATH)/$(GOFILES)
server-build-lin:
	set GOOS=linux
	set GOARCH=amd64
	$(GOBUILD)	-o $(SERVER_PATH)/$(SERVER_BINARY_NAME_LIN) -v $(SERVER_PATH)/$(GOFILES)
server-clean:
	$(GOCLEAN)	$(SERVER_PATH)
build-lin:	client-build-lin	server-build-lin
build-win:	server-build-win	client-build-win	
build-all:	build-lin	build-win
clean:	server-clean	client-clean
test:	
	$(GOTEST)	./...
deps:   
	$(GOGET)	"github.com/spf13/viper"
    
