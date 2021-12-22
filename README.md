# go-tea

Golang entrypoint for REST web apps and client libraries (internal only).

## Installation

go-tea may be installed using the go get command:

```
go get github.com/pghq/go-tea
```
## Usage
To create a new router:

```
import "github.com/pghq/go-tea"

r := tea.NewRouter("")
r.Route("GET", "/test", func(w http.ResponseWriter, r *http.Request){})
```