package srnode

import "testing"

func Test_calcInterfaceUsagePercent(t *testing.T) {
	type args struct {
		firstBytes  int64
		secondBytes int64
		firstTime   int
		secondTime  int
		linkCapBits int64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "correct case",
			args: args{
				firstBytes:  800,
				secondBytes: 1600,
				firstTime:   0,
				secondTime:  10,
				linkCapBits: 8000,
			},
			want: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calcInterfaceUsagePercent(tt.args.firstBytes, tt.args.secondBytes, tt.args.firstTime, tt.args.secondTime, tt.args.linkCapBits); got != tt.want {
				t.Errorf("calcInterfaceUsagePercent() = %v, want %v", got, tt.want)
			}
		})
	}
}
