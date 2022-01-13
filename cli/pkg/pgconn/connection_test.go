package pgconn

import (
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgconn"
)

func TestConstructURI(t *testing.T) {
	tests := []struct {
		user           string
		password       string
		host           string
		port           int
		dbname         string
		sslmode        string
		connectTimeout int
		want           string
	}{
		{
			user:           "bob has spaces in his username",
			password:       "{@nd $p3ci@L ch@R$ 1n h1$ p@$$w0rd!^'",
			host:           "localhost",
			port:           5432,
			dbname:         "bob_has_a_database",
			sslmode:        "",
			connectTimeout: 0,
			want:           "postgres://bob%20has%20spaces%20in%20his%20username:%7B%40nd%20$p3ci%40L%20ch%40R$%201n%20h1$%20p%40$$w0rd%21%5E%27@localhost:5432/bob_has_a_database",
		},
		{
			user:           "$b0b!",
			password:       "$p3ci@L-ch@R$_1n\nh1$	p@$$w0rd!^'",
			host:           "my.host.com",
			port:           9999,
			dbname:         "this-is a valid db nAm3 if double quoted",
			sslmode:        "allow",
			connectTimeout: 10,
			want:           "postgres://$b0b%21:$p3ci%40L-ch%40R$_1n%0Ah1$%09p%40$$w0rd%21%5E%27@my.host.com:9999/this-is%20a%20valid%20db%20nAm3%20if%20double%20quoted?connect_timeout=10&sslmode=allow",
		},
		{
			user:           "postgres",
			password:       "password",
			host:           "my.host.com",
			port:           5555,
			dbname:         "postgres",
			sslmode:        "require", // bot sslmode and connectTimeout
			connectTimeout: 5,
			want:           "postgres://postgres:password@my.host.com:5555/postgres?connect_timeout=5&sslmode=require",
		},
		{
			user:           "postgres",
			password:       "password",
			host:           "my.host.com",
			port:           5555,
			dbname:         "postgres",
			sslmode:        "require",
			connectTimeout: 0, // no connectTimeout
			want:           "postgres://postgres:password@my.host.com:5555/postgres?sslmode=require",
		},
		{
			user:           "postgres",
			password:       "password",
			host:           "my.host.com",
			port:           5555,
			dbname:         "postgres",
			sslmode:        "", // no sslmode
			connectTimeout: 5,
			want:           "postgres://postgres:password@my.host.com:5555/postgres?connect_timeout=5",
		},
		{
			user:           "postgres",
			password:       "", // no password
			host:           "my.host.com",
			port:           5555,
			dbname:         "postgres",
			sslmode:        "",
			connectTimeout: 5,
			want:           "postgres://postgres@my.host.com:5555/postgres?connect_timeout=5",
		},
	}
	for _, tc := range tests {
		connDetails := pgconn.Config{
			Host:           tc.host,
			Port:           uint16(tc.port),
			Database:       tc.dbname,
			User:           tc.user,
			Password:       tc.password,
			ConnectTimeout: time.Second * time.Duration(tc.connectTimeout),
		}
		got := ConstructURI(connDetails, tc.sslmode)
		if tc.want != got {
			t.Errorf("constructURI() got = %v, want %v", got, tc.want)
			return
		}
		if _, err := url.Parse(got); err != nil {
			t.Errorf("constructURI() got = %v, not valid: %v", got, err)
			return
		}
	}
}

func TestUpdatePasswordInDBURI(t *testing.T) {
	type args struct {
		dbURI   string
		newpass string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
		err     error
	}{
		{
			name: "update password in db uri 1",
			args: args{
				dbURI:   "postgres://tsdbadmin:password@hii.hey.com:31822/tsdb?sslmode=require",
				newpass: "password123",
			},
			want:    "postgres://tsdbadmin:password123@hii.hey.com:31822/tsdb?sslmode=require",
			wantErr: false,
		},
		{
			name: "update password in db uri 2",
			args: args{
				dbURI:   "postgres://tsdbadmin:password@hii.hey.com:31822/tsdb?connect_timeout=10&sslmode=require",
				newpass: "password 123",
			},
			want:    "postgres://tsdbadmin:password%20123@hii.hey.com:31822/tsdb?connect_timeout=10&sslmode=require",
			wantErr: false,
		},
		{
			name: "update password in db uri 3",
			args: args{
				dbURI:   "postgres://tsdbadmin:password@hii.hey.com:31822/tsdb?connect_timeout=30&sslmode=require",
				newpass: "pass",
			},
			want:    "postgres://tsdbadmin:pass@hii.hey.com:31822/tsdb?connect_timeout=30&sslmode=require",
			wantErr: false,
		},
		{
			name: "update password in db uri 4",
			args: args{
				dbURI: "postgres://tsdbadmin:password@hii.hey.com:31822/tsdb?connect_timeout=30&sslmode=require",
				newpass: "#$!?()*{^%abcde1234 	.<>'''",
			},
			want:    "postgres://tsdbadmin:%23$%21%3F%28%29%2A%7B%5E%25abcde1234%20%09.%3C%3E%27%27%27@hii.hey.com:31822/tsdb?connect_timeout=30&sslmode=require",
			wantErr: false,
		},
		{
			name: "update password in db uri failure case",
			args: args{
				dbURI:   "@hii.hey.com:31822",
				newpass: "pass",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdatePasswordInDBURI(tt.args.dbURI, tt.args.newpass)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePasswordInDBURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UpdatePasswordInDBURI() got = %v, want %v", got, tt.want)
			}
		})
	}
}
