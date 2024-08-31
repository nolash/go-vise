INPUTS = $(wildcard ./*.vis)
TXTS = $(wildcard ./*.txt.orig)

%.vis:
	go run ../../dev/asm $(basename $@).vis > $(basename $@).bin
	go run ../../dev/gdbm/main.go .

all: $(INPUTS) $(TXTS)

%.txt.orig:
	cp -v $(basename $@).orig $(basename $@)
