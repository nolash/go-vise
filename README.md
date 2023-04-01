# festive: A Constrained Size Output Virtual Machine

An attempt at defining a small VM to create a stack machine for size-constrained clients and servers.

Original motivation was to create a simple templating renderer for USSD clients, combined with an agnostic data-retrieval reference that may conceal any level of complexity.


## Opcodes

The VM defines the following opcode symbols:

* `BACK` - Return to the previous execution frame (will fail if at top frame). It leaves to the state of the execution layer to define what "previous" means.
* `CATCH <symbol> <signal>` - Jump to symbol if signal is set (see `signal` below).
* `CROAK <signal>` - Clear state and restart execution from tip if signal is set (see `signal` below).
* `LOAD <symbol> <size>` - Execute the code symbol `symbol` and cache the data, constrained to the given `size`.
* `RELOAD <symbol>` - Execute a code symbol already loaded by `LOAD` and cache the data, constrained to the given `size`. 
* `MAP <symbol>` - Expose a code symbol previously loaded by `LOAD`. Roughly corresponds to the `global` directive in Python.
* `MOVE <symbol>` - Create a new execution frame, invalidating all previous `MAP` calls.


## Rendering

The fixed-size output is generated using a templating language, and a combination of one or more _max size_ properties, and an optional _sink_ property that will attempt to consume all remaining capacity of the rendered template.

For example, in this example

- `maxOutputSize` is 256 bytes long.
- `template` is 120 bytes long.
- param `one` has max size 10 but uses 5.
- param `two` has max size 20 but uses 12.
- param `three` is a _sink_.

The renderer may use up to `256 - 120 - 5 - 12 = 119` bytes from the _sink_ when rendering the output.
