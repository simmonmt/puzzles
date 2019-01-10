all:
	GOPATH=$$PWD go install ./...

run: all
	bin/vm --ram challenge.bin

clean:
	$(RM) -R bin pkg

FORCE:
