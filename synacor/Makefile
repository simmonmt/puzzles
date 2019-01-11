all:
	GOPATH=$$PWD go install ./...

test:
	GOPATH=$$PWD go test ./...

run: all
	bin/vm --ram challenge.bin

clean:
	$(RM) -R bin pkg

FORCE:
