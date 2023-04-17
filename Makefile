examples: profile session helloworld validate

.PHONY: examples

profile:
	bash examples/compile.bash examples/profile

session:
	bash examples/compile.bash examples/session

helloworld:
	bash examples/compile.bash examples/helloworld

validate:
	bash examples/compile.bash examples/validate
