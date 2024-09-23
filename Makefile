
BINDIR := bin
APP := $(BINDIR)/clich

build:
	go build -o $(APP) ./cmd


# Build and run executables
run: build
	$(APP)

# Delete executable binaries
clean:
	rm -rf $(BINDIR)
