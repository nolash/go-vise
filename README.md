# vise: A Constrained Size Output Virtual Machine

An attempt at defining a small VM to handle menu interaction for size-constrained clients.

Original motivation for this project was to create a simple templating renderer for USSD clients, enhanced with a pluggable and agnostic interface for external data-retrieval.


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
* Dedicated error string to prepend to template (e.g. on catch)


### Pending

* Node Walking Audit Tool (NWAT) to ensure all nodes produce output within constraints.
* Input generator engine for the NWAT.
* State error flag debugger representation, builtin as well as user-defined.
* Stepwise debug view on log/stderr of state mutations.
* Toolset to assist bootstrapping/recovering (failed) state from spec.


### Possibly useful

* Breakpoints.
* Key/value database reference example.


## Documentation

Please refer to `doc/build/index.html`.

Docs can be rebuilt using `make -B doc`.

Documentation sources are found in `doc/texinfo/`


## Examples

Build examples with `make -B examples`
