# go-tea

Golang starter library for REST web apps and client libraries (internal only).

## Installation

go-tea may be installed using the go get command:

```
go get github.com/pghq/go-tea
```
## Usage
To create a new app:

```
import "github.com/pghq/go-tea"

app, err := tea.NewApp()
if err != nil{
    tea.SendError(err)
}
```

To retrieve the router and add routes/middleware:
```
router := app.Router()
router.Get("/hello", func(w http.ResponseWriter, r *http.Request){
    tea.Debug("hello")
    w.Write([]byte("hello"))
})
var middleware []Middleware
http.ListenAndServe(":8080", r.Middleware(middleware...))
```