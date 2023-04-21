# vise: A Constrained Size Output Virtual Machine

An attempt at defining a small VM to handle menu interaction for size-constrained clients and servers.


## Rationale

Original motivation was to create a simple templating renderer for USSD clients, combined with an agnostic data-retrieval reference that may conceal any level of complexity.


## Features

### Implemented

* Define and enforce max output size for every individual output.
* Allow one single data entry to fill remaining available size capacity.
* An assembly-like mini-language to define:
    - external code execution.
    - input validation and routing.
    - menu definitions.
    - flow control.
    - exception handling.
* templated output from results of external code execution.
* generate and navigate pages where data symbol contents are too long to display on a single page.
* pluggable function design for handling external code execution calls.


### Pending

* Node Walking Audit Tool (NWAT) to ensure all nodes produce output within constraints.
* Input generator engine for the NWAT.
* State error flag debugger representation, builtin as well as user-defined.
* Stepwise debug view on log/stderr of state mutations.
* Toolset to assist bootstrapping/recovering (failed) state from spec.


### Possibly useful

* Breakpoints.
* Key/value database reference example.
* Dedicated error string to prepend to template (e.g. on catch)


## Opcodes

The VM defines the following opcode symbols, alphabetically:

* `CATCH <symbol> <signal>` - Jump to symbol if signal is set (see `signals` below). If match, has same side-effect as move.
* `CROAK <signal>` - Clear state and restart execution from top if signal is set (see `signals` below). If match, has same side-effect as move.
* `HALT` - Stop execution. The remaining bytecode (typically, the routing code for the node) is returned to the invoking function.
* `INCMP <arg> <symbol>` - Compare registered input to `arg`. If match, it has the same side-effects as `MOVE`. In addition, any consecutive `INCMP` matches will be ignored until `HALT` is called.
* `LOAD <symbol> <size>` - Execute the code symbol `symbol` and cache the data, constrained to the given `size`. Can be exposed with `MAP` within scope.  See "External code" below.
* `MAP <symbol>` - Expose a code symbol previously loaded by `LOAD` to the rendering client. Roughly corresponds to the `global` directive in Python.
* `MNEXT <choice> <display>` - Define how to display the choice for advancing when browsing menu.
* `MOUT <choice> <display>` - Add menu display entry. Each entry should have a matching `INCMP` whose `arg` matches `choice`. `display` is a descriptive text of the menu item.
* `MOVE <symbol>` - Create a new execution frame, invalidating all previous `MAP` calls.
* `MPREV <choice> <display>` - Define how to display the choice for returning when browsing menu.
* `MSEP` -  **Not yet implemented**. Marker for menu page separation. Incompatible with browseable nodes.
* `MSIZE <max> <min>` - **Not yet implemented**. Set min and max display size of menu part to `num` bytes.
* `RELOAD <symbol>` - Execute a code symbol already loaded by `LOAD` and cache the data, constrained to the previously given `size` for the same symbol. See "External code" below.


### External code

`LOAD` is used to execute code symbols in the host environment. It is loaded with a size constraint, and returned data violating this constraint should generate an error.

Any symbol successfully loaded with `LOAD` will be associated with the call stack frame it is loaded. The symbol will be available in the same frame and frames below it. Once the frame goes out of scope (e.g. `BACK` is called in that frame) the symbols should be freed as soon as possible. At this point they are not available to the abandoned scope.

Loaded symbols are not automatically exposed to the rendering client. To expose symbols ot the rendering client the `MAP` opcode must be used.

The associated content of loaded symbols may be refreshed using the `RELOAD` opcode. `RELOAD` only works within the same constraints as `MAP`. However, updated content must be available even if a `MAP` precedes a `RELOAD` within the same frame.

Methods handling `LOAD` symbols have the client input available to them.


### External symbol optimizations

Only `LOAD` and `RELOAD` should trigger external code side-effects. 

In an effort to prevent leaks from unnecessary external code executions, the following constraints are assumed:

- An explicit `MAP` **must** exist in the scope of any `LOAD`.
- All symbols declared in `MAP` **must** be used for all template renderings of a specific node.

Any code compiler or checked **should** generate an error on any orphaned `LOAD` or `MAP` symbols as described above.


### Signals

Signal may be set when executing of external code symbols, and may be used as a simple exception mechanism.

The signal flag arguments should only set a single flag to be tested. If more than one flag is set, the first flag matched will be used as the trigger.

First 8 flags are reserved and used for internal VM operations.

When a signal is caught, the *bytecode buffer is flushed* before the target symbol code is loaded.


### Avoid duplicate menu items

The vm execution should overwrite duplicate `MOUT` directives with the last definition between `HALT` instructions.

The assembler should detect duplicate `INCMP` and `MOUT` (or menu batch code) selectors, and fail to compile. `MSEP` should be included in duplication detection.


## Menus

A menu has both a display and a input processing part. They are on either side of a `HALT` instruction.

To assist with menu creation, a few batch operation symbols have been made available for use with the assembly language.

* `DOWN <symbol> <choice> <display>` descend to next frame and move to `symbol`
* `UP <choice> <display>` return to the previous frame
* `NEXT <choice> <display>` include pagination advance
* `PREVIOUS <choice> <display>` include pagination return. If `NEXT` has not been defined this will not be rendered.


## Rendering

The fixed-size output is generated using a templating language, and a combination of one or more _max size_ properties, and an optional _sink_ property that will attempt to consume all remaining capacity of the rendered template.

In this example

- `maxOutputSize` is 256 bytes long.
- `template` is 120 bytes long.
- param `one` has max size 10 but uses 5.
- param `two` has max size 20 but uses 12.
- param `three` is a _sink_.
- rendered menu is 15 bytes long.

The renderer may use up to `256 - 120 - 5 - 12 - 15 = 104` bytes from the _sink_ when rendering the output.


### Menu browsing

A max size can be set for the menu, which will count towards the space available for the _template sink_.

Menus too long for a single screen should be browseable through separate screens. How the browse choice is displayed is defined using the `MSEP` definition. The browse choice counts towards the menu size capacity.

When browsing additional menu pages, the template output should not be included.


### Menu defaults

Browsing menu display definitions (`MNEXT`, `MPREV`) as well as size constaints (`MSIZE`) should have sane defaults defined by the assembler if they are missing from the assembly code.


### Multipage support

Multipage outputs, like listings, are handled using the _sink_ output constraints:

- first calculate what the rendered display size is when all symbol results that are _not_ sinks are resolved.
- split and cache the list data within its semantic context, given the _sink_ limitation after rendering.
- provide a `next` and `previous` menu item to browse the prepared pagination of the list data.


### Languages support

Language for rendering can be set as:

* Configuration variable of the engine
* By setting the `FLAG_LANG` flag in the return value of a `LOAD` instruction handler.

The language argument in either case **MUST** be a `ISO-639` language code string. 

Methods for retrieving data and templates are passed the language information through the `context.Context` value under the key `Language`. The value should be cast as a `lang.Language` object.


## Virtual machine interface layout

This is the version `0` of the VM. That translates to  _highly experimental_.

Currently the following rules apply for encoding in version `0`:

- A code instruction is a _big-endian_ 2-byte value. See `vm/opcodes.go` for valid opcode values.
- `symbol` value is encoded as _one byte_ of string length, after which the  byte-value of the string follows.
- `size` value is encoded as _one byte_ of numeric length, after which the _big-endian_ byte-value of the integer follows.
- `signal` value is encoded as _one byte_ of byte length, after which a byte-array representing the defined signal follows.


## Reference implementation

This repository provides a `golang` reference implementation for the `vise` concept.


### Structure

- `asm`: Assembly parser and compiler.
- `cache`: Holds and manages all loaded content.
- `engine`: Outermost interface. Orchestrates execution of bytecode against input. 
- `persist`: Interface and reference implementation of `state` and `cache` persistence across asynchronous vm executions.
- `render`: Renders menu and templates, and enforces output size constraints.
- `resource`: Retrieves data and bytecode from external symbols, and retrieves templates.
- `state`: Holds the bytecode buffer, error states and navigation states.
- `vm`: Defines instructions, and applies transformations according to the instructions.
- `lang`: Validation and specification of language context.


### Template rendering

Template rendering is done using the `text/template` faciilty in the `golang` standard library. 

It expects all replacement symbols to be available at time of rendering, and has no tolerance for missing ones.


## Runtime engine

The runtime engine:

* Validates client input
* Runs VM with client input
* Renders result
* Restarts execution from top if the vm has nothing more to do.

There are two flavors of the engine:

* `engine.Loop` - class used for continuous, in-memory interaction with the vm (e.g. terminal).
* `engine.RunPersisted` - method which combines single vm executions with persisted state (e.g. http).


### Client identification

The `engine.Config` struct defines a property `SessionId` which is added to the `context.Context` passed through entire engine vm call roundtrip.

This is used to identify the caller, and thus defines a top-level storage key under which data entries should be retrieved.


## Bytecode examples

(Minimal, WIP)

```
000a 03666f6f 06746f20666f6f  # MOUT "foo" "to foo" - display a menu entry for choice "foo", described by "to foo"
0008 03666f6f 03626172        # INCMP "foo" "bar"   - move to node "bar" if input is "FOO"
0001 0461696565 01 01         # CATCH "aiee" 1 1    - move to node "aiee" (and immediately halt) if input match flag (1) is not set (1)
0003 04616263 020104          # LOAD "abc" 260      - execute code symbol "abc" with a result size limit of 260 (2 byte BE integer, 0x0104)
0003 04646566 00              # LOAD "def" 0        - execute code symbol "abc" with no size limit (sink)
0005 04616263                 # MAP "abc"           - make "abc" available for renderer
0007                          # HALT                - stop execution (require new input to continue)
0006 03313233                 # MOVE "123"          - move to node "123" (regardless of input)
0007                          # HALT                - stop execution
```

## Assembly examples

See `testdata/*.vis`


## Development tools

Located in the `dev/` directory.


### Test data generation

`go run ./dev/gendata/ <directory>`

Outputs bytecodes and templates for test data scenarios used in `engine` unit tests.


### Interactive runner

`go run ./dev/interactive [-d <data_directory>] [--root <root_symbol>] [--session-id <session_id>] [--persist]`

Creates a new interactive session using `engine.DefaultEngine`, starting execution at symbol `root_symbol`

`data_directory` points to a directory where templates and bytecode is to be found (in the same format as generated by `dev/testdata`).

If `data_directory` is not set, current directory will be used.

if `root_symbol` is not set, the symbol `root` will be used.

if `session_id` is set, mutable data will be stored and retrieved keyed by the given identifer (if implemented).

If `persist` is set, the execution state will be persisted across sessions.


### Assembler

`go run ./dev/asm <assembly_file>`

Will output bytecode on STDOUT generated from a valid assembly file.


### Disassembler

`go run ./dev/disasm/ <binary_file>`

Will list all the instructions on STDOUT from a valid binary file.


## Interactive case examples

Found in `examples/`.

Be sure to `make examples` before running them.

Can be run with e.g. `go run ./examples/<case> [...]`

The available options are the same as for the `dev/interactive` tool.

Contents of the case directory:

* `*.vis` - assembly code.
* `*.bin` - bytecode for each node symbol (only available after make).
* `*.txt.orig` - default contents of a single data entry.
* `*.txt` - current contents of a single data entry (only available after make).
