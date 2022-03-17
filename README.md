# simple-bank-api

This is a simpole bank project, which is a golang web backend learning code.

### note
Please create `.git/hooks/pre-commit`git hook before you need to commit.

```shell
#!/bin/bash

set -ex

if [[ $( git symbolic-ref --short HEAD) = "master" ]]; then
   echo "Please commit with new branch rather than master"
   exit 1
fi

# Filter Glang files match Added (A), Copied (C), Modified (M) conditions.
gofiles=`git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true`

if [ -n "$gofiles" ]; then
    gofmt -s -w $gofiles
    goimports -w $gofiles
    git add $gofiles
fi

golangci-lint run --fix
```
