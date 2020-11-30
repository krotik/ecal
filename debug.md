ECAL Debugger
--
ECAL comes with extensive debugging support featuring:

- Breakpoints


Getting started
--
The simplest way to debug a given program is to run the interpreter in debug mode.

```
ecal debug
```

The interpreter can also start a telnet like debug server.
```
ecal debug -server
```
Note: The debug server is not secured and will run any code which is passed to it.


Debug commands
--
#### `info`
Get environment information.

Example:
```
## info
```

#### `break`
Set a break point to a specific line or identifier.

Parameter | Description
-|-
file and line number as `file:line` / identifier | Line or identifier which should trigger the breakpoint.

Example:
```
## break 5
```

#### `status`
Check all running threads if a breakpoint has been reached and the execution has been halted.

Example:
```
## status
```

#### `inspect`
Show the context of a breakpoint if the execution has been halted.

Parameter | Description
-|-
thread ID | Thread ID of a halted thread.

Example:
```
## inspect 123
```

#### `cont`
Continue the execution of a halted thread.

Parameter | Description
-|-
thread ID | Thread ID of a halted thread.

Example:
```
## cont
```
