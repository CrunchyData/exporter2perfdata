# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
    
BINARY_NAME=pg_metric
TEST_NAME=pg_metric-test.sh

all: build

build: 
	$(GOBUILD) -o $(BINARY_NAME) -v

clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

test:
	./$(TEST_NAME) NODE 1 NODE_1_REPORT.txt
	./$(TEST_NAME) PG   1 PG_1_REPORT.txt

