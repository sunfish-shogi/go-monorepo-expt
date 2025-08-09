GO?=go
BINDIR?=.bin

.PHONY: $(BINDIR)/extradep
$(BINDIR)/extradep:
	mkdir -p $(BINDIR)
	$(GO) build -o $(BINDIR)/extradep main.go

.PHONY: clean
clean:
	rm -rf $(BINDIR)