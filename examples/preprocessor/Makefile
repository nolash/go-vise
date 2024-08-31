INPUTS = $(wildcard ./*.vis)
TXTS = $(wildcard ./*.txt.orig)

%.vis:
	go run ../../dev/asm -f pp.csv $(basename $@).vis > $(basename $@).bin

all: $(INPUTS) $(TXTS)

%.txt.orig:
	cp -v $(basename $@).orig $(basename $@)
