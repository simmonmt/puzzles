all: FORCE
	@root=$$(PWD) ; \
	for i in src/cmd/* ; do \
		cd $$i ; \
		echo $$i ; \
		GOPATH=$$PWD/../../.. go install ; \
		[[ $$? -eq 0 ]] || exit 1 ; \
		cd $$root ; \
	done

run: all
	bin/vm --ram challenge.bin

FORCE:
