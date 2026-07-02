# Request Flow Diagram

This diagram visualizes the lifecycle of a command request in IgnisKV v0.1, starting from user input in the CLI terminal, moving through parsing and execution, and returning the structured output.

---

## 1. ASCII Diagram

```
   [ User Input (CLI Terminal) ]
                 │
                 │  1. Raw string (e.g., "SET user:1 alice\n")
                 ▼
     ┌───────────────────────┐
     │     CLI REPL Loop     │
     └───────────────────────┘
                 │
                 │  2. Read line & strip whitespace
                 ▼
     ┌───────────────────────┐
     │        Parser         │
     └───────────────────────┘
                 │
                 │  3. Tokenizes and constructs Command struct:
                 │     Command{Name: "SET", Args: ["user:1", "alice"]}
                 ▼
     ┌───────────────────────┐
     │      Dispatcher       │
     └───────────────────────┘
                 │
                 │  4. Look up command in Registry.
                 │     Finds SetHandler instance.
                 ▼
     ┌───────────────────────┐
     │    Command Handler    │  (e.g., SetHandler)
     └───────────────────────┘
         │               ▲
         │ 5. Write data │ 6. Go error/values
         ▼               │
     ┌───────────────────────┐
     │      MemoryStore      │  (Encapsulates map[string]Value & RWMutex)
     └───────────────────────┘
         │
         │  7. Returns structured Response struct:
         │     Response{Status: Success, Payload: "OK"}
         ▼
     ┌───────────────────────┐
     │     Output Writer     │  (Part of CLI execution block)
     └───────────────────────┘
                 │
                 │  8. Formats response text
                 ▼
   [ User Output (Console Print) ]
```

---

## 2. Component Pipeline Responsibilities

1. **CLI REPL Loop**: Maintains the interactive execution context, prompts the user, reads standard input, and passes commands to the parser.
2. **Parser**: Implements tokenization logic. In v0.1, it splits by space, respecting quotes if string boundaries are defined, and generates a structured, generic command.
3. **Dispatcher**: Acts as a router. Owns the thread-safe registry mapping command name strings to handlers.
4. **Command Handler**: Validates the command arguments and invokes state-changing methods on the `MemoryStore`.
5. **MemoryStore**: Coordinates thread-safe thread locks (`sync.RWMutex`) and modifies the in-memory map.
6. **Output Writer**: Converts domain values and runtime errors into clean, protocol-appropriate console printouts.
