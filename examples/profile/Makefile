INPUTS = $(wildcard ./*.vis)
TXTS = $(wildcard ./*.txt.orig)

%.vis:
	go run ../../dev/asm $(basename $@).vis > $(basename $@).bin

all: $(INPUTS) $(TXTS)

%.txt.orig:
	cp -v $(basename $@).orig $(basename $@)
