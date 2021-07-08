package pgconn

import (
	"testing"
)

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
				dbURI:   "postgres://tsdbadmin:password@hii.hey.com:31822/tsdb?sslmode=require&connect_timeout=10",
				newpass: "password 123",
			},
			want:    "postgres://tsdbadmin:password 123@hii.hey.com:31822/tsdb?sslmode=require&connect_timeout=10",
			wantErr: false,
		},
		{
			name: "update password in db uri 3",
			args: args{
				dbURI:   "postgres://tsdbadmin:password@hii.hey.com:31822/tsdb?sslmode=require&connect_timeout=30",
				newpass: "pass",
			},
			want:    "postgres://tsdbadmin:pass@hii.hey.com:31822/tsdb?sslmode=require&connect_timeout=30",
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
