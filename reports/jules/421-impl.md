## Verdict
done

The shell lexer has been updated to treat words following `if`, `elif`, `else`, `while`, `until`, and `!` as commands (like it already did for `then` and `do`), by updating the `cmdPos` state in `lexical_shell.go`.

## Changes
- `syntax/lexical_shell.go`: added `!` to `shellKeywords` and updated `cmdPos` trigger words to include `if`, `elif`, `else`, `while`, `until` and `!`.
- `syntax/lexical_shell_test.go`: updated an existing test string regarding `if true; then` and added multiple table tests covering the specified triggers (`if grep`, `while read`, `! test`, `else echo`).

## Tests
I ran the specific tests against the old lexer by extracting them to a temporary file (`syntax/lexical_shell_issue421_test.go`) before modifying the parser to show they were failing:
```
$ go test -run TestShellKeywordsTriggerCommand ./syntax
--- FAIL: TestShellKeywordsTriggerCommand (0.00s)
    --- FAIL: TestShellKeywordsTriggerCommand/if_and_grep (0.00s)
        lexical_shell_issue421_test.go:47: tokens
             got ["keyword:if" "punctuation:;" "keyword:then"]
            want ["keyword:if" "function:grep" "punctuation:;" "keyword:then"]
    --- FAIL: TestShellKeywordsTriggerCommand/while_read (0.00s)
        lexical_shell_issue421_test.go:47: tokens
             got ["keyword:while" "punctuation:;" "keyword:do"]
            want ["keyword:while" "function:read" "punctuation:;" "keyword:do"]
    --- FAIL: TestShellKeywordsTriggerCommand/bang_test (0.00s)
        lexical_shell_issue421_test.go:47: tokens
             got ["function:!"]
            want ["keyword:!" "function:test"]
    --- FAIL: TestShellKeywordsTriggerCommand/else_echo (0.00s)
        lexical_shell_issue421_test.go:47: tokens
             got ["keyword:if" "punctuation:;" "keyword:then" "function:b" "punctuation:;" "keyword:else" "punctuation:;" "keyword:fi"]
            want ["keyword:if" "function:a" "punctuation:;" "keyword:then" "function:b" "punctuation:;" "keyword:else" "function:echo" "punctuation:;" "keyword:fi"]
FAIL
FAIL	ubunatic.com/loom/syntax	0.004s
FAIL
```

After modifying the code and transferring the tests into `syntax/lexical_shell_test.go`:
```
$ go test -run TestShellTokens ./syntax -v
=== RUN   TestShellTokens
...
=== RUN   TestShellTokens/if_and_grep
=== RUN   TestShellTokens/while_read
=== RUN   TestShellTokens/bang_test
=== RUN   TestShellTokens/else_echo
--- PASS: TestShellTokens (0.00s)
...
    --- PASS: TestShellTokens/if_and_grep (0.00s)
    --- PASS: TestShellTokens/while_read (0.00s)
    --- PASS: TestShellTokens/bang_test (0.00s)
    --- PASS: TestShellTokens/else_echo (0.00s)
PASS
ok  	ubunatic.com/loom/syntax	0.005s
```
