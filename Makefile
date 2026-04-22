DESTDIR :=
PREFIX := $(HOME)/.local

.PHONY: all
all:
	go build -o bin/typr src/*.go

.PHONY: clean
clean:
	rm -f bin/typr bin/typr-osx bin/typr.exe bin/typr-linux bin/typr-linux_arm bin/typr-linux_arm64

.PHONY: install
install:
	install -d $(DESTDIR)$(PREFIX)/bin
	install -d $(DESTDIR)$(PREFIX)/share/man/man1
	install -m755 bin/typr $(DESTDIR)$(PREFIX)/bin
	install -m644 typr.1.gz $(DESTDIR)$(PREFIX)/share/man/man1

.PHONY: uninstall
uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/typr
	rm -f $(DESTDIR)$(PREFIX)/share/man/man1/typr.1.gz

.PHONY: assets
assets:
	python3 ./scripts/themegen.py
	./scripts/pack themes/ words/ quotes/ > src/packed.go
	pandoc -s -t man -o - man.md|gzip > typr.1.gz

.PHONY: rel
rel:
	GOOS=darwin GOARCH=amd64 go build -o bin/typr-osx src/*.go
	GOOS=windows GOARCH=amd64 go build -o bin/typr.exe src/*.go
	GOOS=linux GOARCH=amd64 go build -o bin/typr-linux src/*.go
	GOOS=linux GOARCH=arm go build -o bin/typr-linux_arm src/*.go
	GOOS=linux GOARCH=arm64 go build -o bin/typr-linux_arm64 src/*.go
