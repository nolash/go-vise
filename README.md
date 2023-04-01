# festive: A Constrained Size Output Virtual Machine

An attempt at defining a small VM to create a stack machine for size-constrained clients and servers.

Original motivation was to create a simple templating renderer for USSD clients, combined with an agnostic data-retrieval reference that may conceal any level of complexity.


## Opcodes

The VM defines the following opcode symbols:

* `BACK` - Return to the previous execution frame (will fail if at top frame). It leaves to the state of the execution layer to define what "previous" means.
* `CATCH <symbol> <signal>` - Jump to symbol if signal is set (see `signal` below).
* `CROAK <signal>` - Clear state and restart execution from top if signal is set (see `signal` below).
* `LOAD <symbol> <size>` - Execute the code symbol `symbol` and cache the data, constrained to the given `size`.
* `RELOAD <symbol>` - Execute a code symbol already loaded by `LOAD` and cache the data, constrained to the previously given `size` for the same symbol. 
* `MAP <symbol>` - Expose a code symbol previously loaded by `LOAD` to the rendering client. Roughly corresponds to the `global` directive in Python.
* `MOVE <symbol>` - Create a new execution frame, invalidating all previous `MAP` calls. More detailed: After a `MOVE` call, a `BACK` call will return to the same execution frame, with the same symbols available, but all `MAP` calls will have to be repeated.


## Rendering

The fixed-size output is generated using a templating language, and a combination of one or more _max size_ properties, and an optional _sink_ property that will attempt to consume all remaining capacity of the rendered template.

For example, in this example

- `maxOutputSize` is 256 bytes long.
- `template` is 120 bytes long.
- param `one` has max size 10 but uses 5.
- param `two` has max size 20 but uses 12.
- param `three` is a _sink_.

The renderer may use up to `256 - 120 - 5 - 12 = 119` bytes from the _sink_ when rendering the output.


### Multipage support

Multipage outputs, like listings, are handled using the _sink_ output constraints:

- first calculate what the rendered display size is when all symbol results that are _not_ sinks are resolved.
- split and cache the list data within its semantic context, given the _sink_ limitation after rendering.
- provide a `next` and `previous` menu item to browse the prepared pagination of the list data.


## Virtual machine interface layout

This is the version `0` of the VM. That translates to  _highly experimental_.

Currently the following rules apply for encoding in version `0`:

- A code instruction is a _big-endian_ 2-byte value. See `vm/opcodes.go` for valid opcode values.
- `symbol` value is encoded as _one byte_ of string length, after which the  byte-value of the string follows.
- `size` value is encoded as _one byte_ of numeric length, after which the _big-endian_ byte-value of the integer follows.
- `signal` value is encoded as _one byte_ of byte length, after which a byte-array representing the defined signal follows.


## Reference implementation

This repository provides a `golang` reference implementation for the `festive` concept.

In this reference implementation some constraints apply


### Template rendering

Template rendering is done using the `text/template` faciilty in the `golang` standard library. 

It expects all replacement symbols to be available at time of rendering, and has no tolerance for missing ones.


## Assembly language

**TBD**

An assmebly language will be defined to generate the _routing_ and _execution_ bytecodes for each menu node.
