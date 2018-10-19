GOCMD   =  go
GOBUILD =  $(GOCMD)  build
GOINSTALL =  $(GOCMD)  install
GOCLEAN =  $(GOCMD)    clean
GOTEST  =  $(GOCMD)    test
GOGET   =  $(GOCMD)    get
GOSET   =              set
BINARY_NAME_CLINET_WIN =  client.exe
BINARY_NAME_CLIENT_LIN =  client_lin


clibuildwin:	
	$(GOBUILD)	-o $(BINARY_NAME_CLINET_WIN) -v client.go	clientconfig.go
test:	
	$(GOTEST)	./...
clirunwin:	clibuildwin
	./$(BINARY_NAME_CLINET_WIN)
	$(GOCLEAN)

# run:    
#     nmake windowsbuild    
#     $(BINARY_NAME)
# deps:   
#     $(GOGET)    "github.com/spf13/viper"
# clean:    
#     $(GOCLEAN) -a
#     del  $(BINARY_NAME)

# inst:    
#     $(GOINSTALL) -i

# linuxbuild:
#     nmake proto    
#     SET GOOS=linux
#     SET GOARCH=amd64
#     $(GOBUILD) -o $(BINARY_NAME_LINUX)
    
# dockerbuild:
#     nmake linuxbuild
#     docker build -t vahidjafari/unitservice:1.0 .

# dockerrun:
#     docker run -ti --rm vahidjafari/unitservice:1.0
