package cmd

import (
	"reflect"
	"testing"
)

func Test_makeSIDList(t *testing.T) {
	type args struct {
		srcAddr          string
		dstAddr          string
		topologyFilePath string
	}
	tests := []struct {
		name    string
		args    args
		want    *[]string
		wantErr bool
	}{
		{
			name: "normal case",
			args: args{
				srcAddr:          "fd00:0:0:1::1",
				dstAddr:          "fd00:0:0:5::1",
				topologyFilePath: "../testdata/test1.txt",
			},
			want:    &[]string{"fd00:0:0:1::1", "fd00:0:0:3::1", "fd00:0:0:2::1", "fd00:0:0:5::1"},
			wantErr: false,
		}, {
			name: "empty source address",
			args: args{
				srcAddr:          "",
				dstAddr:          "fd00:0:0:5::1",
				topologyFilePath: "../testdata/test1.txt",
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "empty destination address",
			args: args{
				srcAddr:          "fd00:0:0:1::1",
				dstAddr:          "",
				topologyFilePath: "../testdata/test1.txt",
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "no path from source to destination",
			args: args{
				srcAddr:          "fd00:0:0:1::2",
				dstAddr:          "fd00:0:0:1::5",
				topologyFilePath: "../testdata/test1.txt",
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "no key that matches source address",
			args: args{
				srcAddr:          "fd00:0:0:2::1",
				dstAddr:          "fd00:0:0:1::5",
				topologyFilePath: "../testdata/test1.txt",
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "no key that matches destination address",
			args: args{
				srcAddr:          "fd00:0:0:1::1",
				dstAddr:          "fd00:0:0:2::5",
				topologyFilePath: "../testdata/test1.txt",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeSIDList(tt.args.srcAddr, tt.args.dstAddr, tt.args.topologyFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeSIDList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeSIDList() got = %v, want %v", got, tt.want)
			}
		})
	}
}
