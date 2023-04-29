all: profile session helloworld validate

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
