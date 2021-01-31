package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/vishvananda/netlink"

	"github.com/vishvananda/netns"

	"github.com/chez-shanpu/acar/api"
	"github.com/chez-shanpu/acar/api/types"
)

type tearDownNetlinkTest func()

func Test_dataplaneServer_ApplySRPolicy(t *testing.T) {
	type fields struct {
		UnimplementedDataPlaneServer api.UnimplementedDataPlaneServer
		devName                      string
	}
	type args struct {
		ctx context.Context
		si  *api.SRInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.Result
		wantErr bool
	}{
		{
			name: "correct case (IPv4)",
			fields: fields{
				UnimplementedDataPlaneServer: api.UnimplementedDataPlaneServer{},
				devName:                      "lo",
			},
			args: args{
				ctx: context.Background(),
				si: &api.SRInfo{
					SrcAddr: "192.168.1.1/32",
					DstAddr: "192.168.2.1/32",
					SidList: []string{"fd00:0:0:1::1", "fd00:0:0:3::1", "fd00:0:0:2::1", "fd00:0:0:5::1"},
				},
			},
			want: &types.Result{
				Ok:     true,
				ErrStr: "",
			},
			wantErr: false,
		}, {
			name: "correct case (IPv6)",
			fields: fields{
				UnimplementedDataPlaneServer: api.UnimplementedDataPlaneServer{},
				devName:                      "lo",
			},
			args: args{
				ctx: context.Background(),
				si: &api.SRInfo{
					SrcAddr: "2001:0:0:10::1/64",
					DstAddr: "2001:0:0:20::1/64",
					SidList: []string{"fd00:0:0:1::1", "fd00:0:0:3::1", "fd00:0:0:2::1", "fd00:0:0:5::1"},
				},
			},
			want: &types.Result{
				Ok:     true,
				ErrStr: "",
			},
			wantErr: false,
		}, {
			name: "empty sid list",
			fields: fields{
				UnimplementedDataPlaneServer: api.UnimplementedDataPlaneServer{},
				devName:                      "lo",
			},
			args: args{
				ctx: context.Background(),
				si: &api.SRInfo{
					SrcAddr: "2001:0:0:10::1/64",
					DstAddr: "2001:0:0:20::1/64",
					SidList: nil,
				},
			},
			want: &types.Result{
				Ok:     false,
				ErrStr: "",
			},
			wantErr: true,
		}, {
			name: "empty source address",
			fields: fields{
				UnimplementedDataPlaneServer: api.UnimplementedDataPlaneServer{},
				devName:                      "lo",
			},
			args: args{
				ctx: context.Background(),
				si: &api.SRInfo{
					SrcAddr: "",
					DstAddr: "2001:0:0:20::1/64",
					SidList: []string{"fd00:0:0:1::1", "fd00:0:0:3::1", "fd00:0:0:2::1", "fd00:0:0:5::1"},
				},
			},
			want: &types.Result{
				Ok:     false,
				ErrStr: "",
			},
			wantErr: true,
		}, {
			name: "empty destination address",
			fields: fields{
				UnimplementedDataPlaneServer: api.UnimplementedDataPlaneServer{},
				devName:                      "lo",
			},
			args: args{
				ctx: context.Background(),
				si: &api.SRInfo{
					SrcAddr: "2001:0:0:10::1/64",
					DstAddr: "",
					SidList: []string{"fd00:0:0:1::1", "fd00:0:0:3::1", "fd00:0:0:2::1", "fd00:0:0:5::1"},
				},
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "empty device name",
			fields: fields{
				UnimplementedDataPlaneServer: api.UnimplementedDataPlaneServer{},
				devName:                      "",
			},
			args: args{
				ctx: context.Background(),
				si: &api.SRInfo{
					SrcAddr: "",
					DstAddr: "",
					SidList: nil,
				},
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "wrong format source address",
			fields: fields{
				UnimplementedDataPlaneServer: api.UnimplementedDataPlaneServer{},
				devName:                      "lo",
			},
			args: args{
				ctx: context.Background(),
				si: &api.SRInfo{
					SrcAddr: "A",
					DstAddr: "2001:0:0:20::1/64",
					SidList: []string{"fd00:0:0:1::1", "fd00:0:0:3::1", "fd00:0:0:2::1", "fd00:0:0:5::1"},
				},
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "wrong format destination address",
			fields: fields{
				UnimplementedDataPlaneServer: api.UnimplementedDataPlaneServer{},
				devName:                      "lo",
			},
			args: args{
				ctx: context.Background(),
				si: &api.SRInfo{
					SrcAddr: "2001:0:0:10::1/64",
					DstAddr: "B",
					SidList: []string{"fd00:0:0:1::1", "fd00:0:0:3::1", "fd00:0:0:2::1", "fd00:0:0:5::1"},
				},
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "wrong device name",
			fields: fields{
				UnimplementedDataPlaneServer: api.UnimplementedDataPlaneServer{},
				devName:                      "hoge",
			},
			args: args{
				ctx: context.Background(),
				si: &api.SRInfo{
					SrcAddr: "2001:0:0:10::1/64",
					DstAddr: "2001:0:0:20::1/64",
					SidList: []string{"fd00:0:0:1::1", "fd00:0:0:3::1", "fd00:0:0:2::1", "fd00:0:0:5::1"},
				},
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "IPv4 source address & IPv6 destination address",
			fields: fields{
				UnimplementedDataPlaneServer: api.UnimplementedDataPlaneServer{},
				devName:                      "lo",
			},
			args: args{
				ctx: context.Background(),
				si: &api.SRInfo{
					SrcAddr: "192.168.1.1/32",
					DstAddr: "2001:0:0:10::1/64",
					SidList: []string{"fd00:0:0:1::1", "fd00:0:0:3::1", "fd00:0:0:2::1", "fd00:0:0:5::1"},
				},
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "IPv6 source address & IPv4 destination address",
			fields: fields{
				UnimplementedDataPlaneServer: api.UnimplementedDataPlaneServer{},
				devName:                      "lo",
			},
			args: args{
				ctx: context.Background(),
				si: &api.SRInfo{
					SrcAddr: "2001:0:0:20::1/64",
					DstAddr: "192.168.1.1/32",
					SidList: []string{"fd00:0:0:1::1", "fd00:0:0:3::1", "fd00:0:0:2::1", "fd00:0:0:5::1"},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	if os.Getuid() != 0 {
		t.Skip("Test requires root privileges.")
	}

	// todo if exec `runtime.LockOSThread()` host network rule is polluted
	//runtime.LockOSThread()
	ns, err := netns.New()

	if err != nil {
		t.Fatal("Failed to create newns", ns)
	}
	link, err := netlink.LinkByName("lo")
	if err != nil {
		t.Fatal(err)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = ns.Close()
		//runtime.UnlockOSThread()
	}()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &dataplaneServer{
				UnimplementedDataPlaneServer: tt.fields.UnimplementedDataPlaneServer,
				devName:                      tt.fields.devName,
			}
			got, err := s.ApplySRPolicy(tt.args.ctx, tt.args.si)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplySRPolicy() error = %v, wantErr: %v, got: %v", err, tt.wantErr, got)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ApplySRPolicy() got = %v, want %v", got, tt.want)
			//}
			//if tt.wantErr == false {
			//	routesV6, err := netlink.RouteList(link, netlink.FAMILY_V6)
			//	if err != nil {
			//		t.Fatal(err)
			//	}
			//	routesV4, err := netlink.RouteList(link, netlink.FAMILY_V4)
			//	if err != nil {
			//		t.Fatal(err)
			//	}
			//
			//	if len(routesV4) != 1 && len(routesV6) != 1 {
			//		t.Fatal("SR-policy is not added properly")
			//	}
			//}
		})
	}
}
