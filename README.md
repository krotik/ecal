ECAL
====

<p align="center">
  <img height="150px" style="height:150px;" src="ecal-support/images/logo.png">
</p>

ECAL is an ECA (Event Condition Action) language for concurrent event processing. ECAL can define event-based systems using rules which are triggered by events. ECAL is intended to be embedded into other software to provide an easy to use scripting language which can react to external events.

Features
--------
- Simple intuitive syntax
- Minimalistic base language (by default only writing to a log is supported)
- Language can be easily extended either by auto generating bridge adapters to Go functions or by adding custom function into the stdlib
- External events can be easily pushed into the interpreter and scripts written in ECAL can react to these events.
- Simple but powerful concurrent event-based processing supporting priorities and scoping for control flow.
- Handling event rules can match on event state and rules can suppress each other.

### Getting started

Clone the repository and build the ECAL executable with a simple `make` command. You need Go 1.14 or higher.

Run `./ecal` to start an interactive session. You can now write simple one line statements and evaluate them:

```
>>>a:=2;b:=a*4;a+b
10
>>>"Result is {{a+b}}"
Result is 10
```

Close the interpreter by pressing <ctrl>+d and change into the directory `examples/fib`.
There are 2 ECAL files in here:

lib.ecal
```
# Library for fib

/*
fib calculates the fibonacci series using recursion.
*/
func fib(n) {
    if (n <= 1) {
        return n
    }
    return fib(n-1) + fib(n-2)
}
```

fib.ecal
```
import "lib.ecal" as lib

for a in range(2, 20, 2) {
  log("fib({{a}}) = ", lib.fib(a))
}
```

Run the ECAL program with: `sh run.sh`. The output should be like:
```
$ sh run.sh
2000/01/01 12:12:01 fib(2) = 1
2000/01/01 12:12:01 fib(4) = 3
2000/01/01 12:12:01 fib(6) = 8
2000/01/01 12:12:01 fib(8) = 21
2000/01/01 12:12:01 fib(10) = 55
2000/01/01 12:12:01 fib(12) = 144
2000/01/01 12:12:02 fib(14) = 377
2000/01/01 12:12:02 fib(16) = 987
2000/01/01 12:12:02 fib(18) = 2584
2000/01/01 12:12:02 fib(20) = 6765
```

The interpreter can be run in debug mode which adds debug commands to the console. Run the ECAL program in debug mode with: `sh debug.sh` - this will also start a debug server which external development environments can connect to. There is a [VSCode integration](ecal-support/README.md) available which allows debugging via a graphical interface.

### Further Reading:

- [ECA Language](ecal.md)
- [ECA Engine](engine.md)
- [VSCode integration](ecal-support/README.md)

License
-------
ECAL source code is available under the [MIT License](/LICENSE).
