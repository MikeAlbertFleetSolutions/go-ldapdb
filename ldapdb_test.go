package ldapdb

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"gopkg.in/ldap.v2"
	"gopkg.in/yaml.v2"
)

// configuration holds the test configuration
type configuration struct {
	LdapConnection struct {
		ServerPort string
		Username   string
		Password   string
	}
	TestConnQuery struct {
		ValidQuery          string
		ValidQueryWithWhere string
		ValidQueryNoWhere   string
	}
}

var (
	config configuration
)

func init() {
	// read config
	// #nosec G304
	bytes, err := ioutil.ReadFile("ldapdb_test.yaml")
	if err != nil {
		log.Fatalf("%+v", err)
	}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func getTestLdapClient() (client *ldap.Conn, err error) {
	// connect to ldap server
	client, err = ldap.Dial("tcp", config.LdapConnection.ServerPort)
	if err != nil {
		return
	}

	// TLS upgrade, #nosec G402
	err = client.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		client = nil
		return
	}

	// authenticate to ldap server
	err = client.Bind(config.LdapConnection.Username, config.LdapConnection.Password)
	if err != nil {
		client = nil
		return
	}

	return
}

func TestGetResults(t *testing.T) {
	connString := fmt.Sprintf("%s/%s@%s", config.LdapConnection.Username, config.LdapConnection.Password, config.LdapConnection.ServerPort)

	conn, err := sql.Open("ldapdb", connString)
	if err != nil {
		t.Fatalf("sql.Open error: %+v", err)
	}

	rows, err := conn.Query(config.TestConnQuery.ValidQuery)
	if err != nil {
		t.Fatalf("conn.Query error: %+v", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("rows.Columns error: %+v", err)
	}
	if len(cols) == 0 {
		t.Fatalf("no columns returned")
	}

	//log.Println(cols)

	// go thru this just to validate pointers, etc.
	vals := make([]interface{}, len(cols))
	for i := range vals {
		vals[i] = new(interface{})
	}
	strVals := make([]string, len(cols))

	numRows := 0
	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			t.Fatalf("rows.Scan error: %+v", err)
		}

		for i, val := range vals {
			if *(val.(*interface{})) != nil {
				strVals[i] = fmt.Sprintf("%v", *(val.(*interface{})))
			} else {
				strVals[i] = ""
			}
		}

		//log.Println(strVals)

		numRows++
	}

	if numRows == 0 {
		t.Fatalf("no rows returned")
	}
}
