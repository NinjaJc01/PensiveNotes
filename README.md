# PensiveNotes
PensiveNotes is a note taking app for those who want to think and reflect about what they write later on.

## Dependencies
github.com/gorilla/mux
github.com/mattn/go-sqlite3

## Installation
Install golang (1.14+)
Clone the repository
`go get github.com/mattn/go-sqlite3`
`go get github.com/gorilla/mux`
`go build -o server main.go`
`./server`
If you'd like the server to be able to bind to ports under 1000, I recommend using Linux capabilities rather than running as root.

## Using PensiveNotes
After downloading and compiling PensiveNotes, log in using the default credentials `pensive:PensiveNotes`

Make sure you change this password immediately!