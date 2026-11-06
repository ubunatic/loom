## Verdict
done

I extracted `nextClusterBoundary` into a shared function for `StringWidthNew` and `ClustersNew`, and added a test showing that an isolated VS16 fails in the old implementation but passes in the new one.

## Changes
measure/measure.go
measure/measure_parity_test.go

## Tests
`git stash`
```
Saved working directory and index state WIP on jules-2652737723992373498-569c5c20: 2265d02 docs(issues): 429 Jules wave 1 tracking (taia focus)
```

`go test -run TestClustersWidthParity ./measure`
```
--- FAIL: TestClustersWidthParity (0.00s)
    measure_parity_test.go:22: Clusters width mismatch for "️": sum=2, StringWidthNew=0
FAIL
FAIL	ubunatic.com/loom/measure	0.009s
FAIL
```

`git stash pop`
```
On branch jules-2652737723992373498-569c5c20
Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
	modified:   measure/measure.go

Untracked files:
  (use "git add <file>..." to include in what will be committed)
	measure/measure_parity_test.go

no changes added to commit (use "git add" and/or "git commit -a")
Dropped refs/stash@{0} (aa0f3a11c4f0e1409be1c079958c2c8d06ed9fe9)
```

`go test -run TestClustersWidthParity ./measure`
```
ok  	ubunatic.com/loom/measure	0.008s
```
