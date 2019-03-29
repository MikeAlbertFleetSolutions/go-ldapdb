package ldapdb

import (
	"fmt"
	"testing"
)

func TestDriver_Open(t *testing.T) {
	type args struct {
		conString string
	}
	tests := []struct {
		name     string
		d        ldapdriver
		args     args
		wantConn bool
		wantErr  bool
	}{
		{
			"Valid connection parameters",
			ldapdriver{},
			args{fmt.Sprintf("%s/%s@%s", config.LdapConnection.Username, config.LdapConnection.Password, config.LdapConnection.ServerPort)},
			true,
			false,
		},
		{
			"Invalid connection string",
			ldapdriver{},
			args{"butchered-connection-string"},
			false,
			true,
		},
		{
			"Valid connection but can't connection",
			ldapdriver{},
			args{"username/password@server:port"},
			false,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ldapdriver{}
			gotConn, err := d.Open(tt.args.conString)
			if (err != nil) != tt.wantErr {
				t.Errorf("Driver.Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotConn != nil) != tt.wantConn {
				t.Errorf("Driver.Open() = %v, want %v", gotConn, tt.wantConn)
			}
		})
	}
}
