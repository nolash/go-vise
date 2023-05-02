all: bin examples doc

examples: profile session helloworld validate intro

bin:
	mkdir -p build
	go build -o build/interactive ./dev/interactive
	go build -o build/gendata ./dev/gendata
	go build -o build/asm ./dev/asm
	go build -o build/disasm ./dev/disasm

profile:
	make -C examples/profile

session:
	make -C examples/session

helloworld:
	make -C examples/session

validate:
	make -C examples/validate

longmenu:
	make -C examples/longmenu

intro:
	make -C examples/intro

doc:
	make -C doc
