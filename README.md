# go-ldapdb
Golang LDAP database driver conforming to the Go database/sql interface.

## Installation
You need a working environment with Go installed and $GOPATH set.

Download and install the ldap database/sql driver:
```bash
go get github.com/MikeAlbertFleetSolutions/go-ldapdb
```
Make sure you have Git installed and in your $PATH.

## Usage
This is an implementation of Go's database/sql/driver interface. In order to use it, you need to import the package and use the database/sql API.

Only SELECT operations are supported.

```go
import (
	"database/sql"
	"log"

	_ "github.com/MikeAlbertFleetSolutions/go-ldapdb"
)

func main() {
	conn, err := sql.Open("ldapdb", "username/password@server:port")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	rows, err := conn.Query("select cn from OU=Users,OU=Corporate AD,DC=company,DC=corp where (objectCategory=user)")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	...

}
```
