package ldapdb

import (
	"database/sql/driver"
	"testing"
)

func Test_conn_Query(t *testing.T) {
	type fields struct {
	}
	type args struct {
		query string
		args  []driver.Value
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantRows bool
		wantErr  bool
	}{
		{
			"Test valid statement with where clause",
			fields{},
			args{config.TestConnQuery.ValidQueryWithWhere, nil},
			true,
			false,
		},
		{
			"Test valid statement without where clause",
			fields{},
			args{config.TestConnQuery.ValidQueryNoWhere, nil},
			true,
			false,
		},
		{
			"Test bad statement",
			fields{},
			args{"bad-ldap-query", nil},
			false,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := getTestLdapClient()
			if err != nil {
				t.Errorf("can't establish ldap connection for test")
				return
			}

			c := &Conn{
				client: l,
			}
			gotRows, err := c.Query(tt.args.query, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("conn.Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotRows != nil) != tt.wantRows {
				t.Errorf("conn.Query() = %v, want %v", gotRows, tt.wantRows)
			}
		})
	}
}
