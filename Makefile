BUILDDIR := build
BINARY := jardb
PLATFORMS := linux/386 linux/amd64 windows/386/.exe windows/amd64/.exe
VERSION ?= dev

OS = $(word 1, $(subst /, ,$@))
ARCH = $(word 2, $(subst /, ,$@))
EXT = $(word 3, $(subst /, ,$@))
OUTFILE = $(BUILDDIR)/$(BINARY)-$(VERSION)-$(OS)-$(ARCH)$(EXT)

.PHONY: jardb
jardb:
	go build -o $(BUILDDIR)/$(BINARY)

.PHONY: clean
clean:
	rm -rf $(BUILDDIR)

.PHONY: install
install: jardb
	cp $(BUILDDIR)/$(BINARY) $(GOPATH)/bin

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUTFILE)

release: $(PLATFORMS)
