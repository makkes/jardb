.PHONY: jardb
jardb:
	go build -o build/jardb

.PHONY: clean
clean:
	rm -rf build

.PHONY: install
install: jardb
	cp build/jardb $(GOPATH)/bin
