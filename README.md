# vise: A Constrained Size Output Virtual Machine

Consider the following interaction:


```
This is the root page
You have visited 1 time.
0:foo
1:bar

$ 1
Please visit foo first.
Any input to return.

$ x
This is the root page.
You have visited 2 times.
0:foo
1:bar

$ 0
Welcome to page foo.
Please write seomthing.

$ blah blah blah
This is the root page.
You have visited 3 times.
0:foo
1:bar

$ 1
Thanks for visiting foo and bar.
You have written:
blah blah blah
```

## Components

This simple interface above involves four different menu nodes.

In order to engineer these using vise, three types of components are involved:

* An assembly-like menu handling script.
* A display template.
* External code handlers for the counter and the "something" input.


## Nodes

* `root` - The first page
* `foo` - The "foo" page
* `bar` - The "bar" page after "foo" has been visited
* `ouch` - The "bar" page before "foo" has been visited


## Templates
Each page has a template that may or may not contain dynamic elements.

In this example the root and bar nodes contains dynamic content.

### root

```
This is the root page
You have visited {{.count}}.
```


### foo

```
Welcome to page foo.
Please write something.
```


### bar

```
Thanks for visiting foo and bar.
You have written:
{{.something}}
```


### ouch

```
Please visit foo first.
Any input to return.
```


## Scripts

The scripts are responsible for defining menus, handling navigation flow control, and triggering external code handlers.

### root

```
LOAD count 8        # trigger external code handler "count"
LOAD something 0    # trigger external code handler "something"
RELOAD count        # explicitly trigger "count" every time this code is executed.
MAP count           # make the result from "count" available to the template renderer
MOUT foo 0          # menu item
MOUT bar 1          # menu item
HALT                # render template and wait for input
INCMP foo 0         # match menu selection 0, move to node "foo" on match
INCMP bar 1         # match menu selection 1, move to node "bar" on match
```

### foo

```
HALT                # render template and wait for input
RELOAD something    # pass input to the "something" external code handler.
                    # The input will be appended to the stored value. 
                    # The "HAVESOMETHING" flag (8) will be set.
MOVE _              # move up one level
```


### bar

```
CATCH ouch 8 0      # if the "HAVESOMETHING" (8) flag has NOT (0) been set, move to "ouch"
MNEXT next 11       # menu choice to display for advancing one page
MPREV back 22       # menu choice to display for going back to the previous page
MAP something       # make the result from "something" available to the template renderer
HALT                # render template and wait for input
INCMP > 11          # handle the "next" menu choice
INCMP < 22          # handle to "back" menu choice
INCMP _ *           # move to the root node on any input
```


### ouch

```
HALT            # render template and wait for input
INCMP ^ *       # move to the root node on any input
```

## External code handlers

The script code contains `LOAD` instructions for two different methods. 

```
import (
    "context"
    "fmt"
    "path"
    "strings"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/resource"
)

const (
	USERFLAG_HAVESOMETHING = iota + state.FLAG_USERSTART
)

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "intro")
)

type introResource struct {
	*resource.FsResource 
	c int64
	v []string
}

func newintroResource() introResource {
	fs := resource.NewFsResource(scriptDir)
	return introResource{fs, 0, []string{}}
}

// increment counter.
// return a string representing the current value of the counter.
func(c *introResource) count(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	s := "%v time"
	if c.c != 1 {
		s += "s"
	}
	r := resource.Result{
		Content: fmt.Sprintf(s, c.c),
	}
	c.c += 1 
	return  r, nil
}

// if input is suppled, append it to the stored string vector and set the HAVESOMETHING flag.
// return the stored string vector value, one string per line.
func(c *introResource) something(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	c.v = append(c.v, string(input))
	r := resource.Result{
		Content: strings.Join(c.v, "\n"),
	}
	if len(input) > 0 {
		r.FlagSet = []uint32{USERFLAG_HAVESOMETHING}
	}
	return r, nil
}
```

## Handling long values

In the above example, the more times the `foo` page is supplied with a value, the longer the vector of values that need to be displayed by the `bar` page will be.

A core feature of `vise` is to magically create browseable pages from these values from a pre-defined maximum output capacity for each page.

Consider the case where the contents of the `something` symbol has become:

```
foo bar
baz bazbaz
inky pinky
blinky
clyde
```

Given a size constaint of 90 characters, the display will be split into two pages:

```
Thanks for visiting foo and bar.
You have written:
foo bar
baz bazbaz
11:next
```

```
Thanks for visiting foo and bar.
You have written:
inky pinky
blinky
clyde
22:back
```


## Working example

In the source code repository, a full working example of this menu can be found in `examples/intro`.

To run it:

```
make -B intro
go run ./examples/intro
```

Use `go run -tags logtrace ...` to peek at what is going on under the hood.

To play the "Handling long values" case above, limit the output size by adding `-s 90`.
