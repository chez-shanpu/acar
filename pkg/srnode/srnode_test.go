package srnode

import "testing"

func Test_calcInterfaceUsagePercent(t *testing.T) {
	type args struct {
		firstBytes  int64
		secondBytes int64
		duration    float64
		linkCapBits int64
	}
	tests := []struct {
		name  string
		args  args
		want1 float64
		want2 float64
	}{
		{
			name: "correct case",
			args: args{
				firstBytes:  800,
				secondBytes: 1600,
				duration:    10,
				linkCapBits: 10000,
			},
			want1: 6.4,
			want2: 80,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, got2 := calcInterfaceUsage(tt.args.firstBytes, tt.args.secondBytes, tt.args.duration, tt.args.linkCapBits)
			if got1 != tt.want1 {
				t.Errorf("calcInterfaceUsage() got = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("calcInterfaceUsage() got1 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
