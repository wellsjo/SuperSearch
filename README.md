# SuperSearch

Attempting to create a faster grep tool in Go.

## Install
Install everything with `make`.

## Usage
```
ss [OPTIONS] PATTERN [PATH]

Application Options:
  -q, --quiet         Doesn't log any matches, just the results summary
      --hidden        Search hidden files
  -U, --unrestricted  Search all files (ignore .gitignore)
  -D, --debug         Show verbose debug information

Help Options:
  -h, --help          Show this help message
```
