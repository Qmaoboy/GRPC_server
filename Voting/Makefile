CC=protoc

FLAG1=--go_out=
FLAG2=--go-grpc_out=
PROGS= *.proto
DIRNAME=Voting
FILE_DIR=$(GOROOT)/src/$(DIRNAME)

.PHONY:all
all: $(PROGS)
	@$(CC) $(FLAG1). $(FLAG2). $(PROGS)
	@echo "Update File to" $(shell pwd)

goabse:
	@$(CC) $(FLAG1)$(FILE_DIR) $(FLAG2)$(FILE_DIR) $(PROGS)
	@echo "Update File to"$(GOROOT)


### protoc --go_out=. --go-grpc_out=. *.proto
