# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
    
BINARY_NAME=exporter2perfdata
TEST_NAME=testing/exporter2perfdata-test.sh

all: build

build: 
	CGO_ENABLED=0 $(GOBUILD) -o $(BINARY_NAME) -v

clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

test:
	./$(TEST_NAME) NODE 1
	./$(TEST_NAME) PG   1
	./$(TEST_NAME) NODE 2
	./$(TEST_NAME) PG   2

testclean:
	rm -f testing/NODE_[0-9]*_REPORT.tmp
	rm -f testing/PG_[0-9]*_REPORT.tmp
	rm testing/cmd.sh testing/debug.tmp

