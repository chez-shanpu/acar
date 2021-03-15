package appagent

import (
	"reflect"
	"testing"

	"github.com/chez-shanpu/acar/api"

	"github.com/RyanCarrier/dijkstra"
)

const testFileDir = "../../testdata/appagent/"

func Test_makeSIDList(t *testing.T) {
	type args struct {
		srcAddr          []string
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
				srcAddr:          []string{"fd00:0:0:1::1"},
				dstAddr:          "fd00:0:0:5::1",
				topologyFilePath: testFileDir + "test1.txt",
			},
			want:    &[]string{"fd00:0:0:1::1", "fd00:0:0:3::1", "fd00:0:0:2::1", "fd00:0:0:5::1"},
			wantErr: false,
		}, {
			name: "empty source address",
			args: args{
				srcAddr:          []string{""},
				dstAddr:          "fd00:0:0:5::1",
				topologyFilePath: testFileDir + "test1.txt",
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "empty destination address",
			args: args{
				srcAddr:          []string{"fd00:0:0:1::1"},
				dstAddr:          "",
				topologyFilePath: testFileDir + "test1.txt",
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "no path from source to destination",
			args: args{
				srcAddr:          []string{"fd00:0:0:1::2"},
				dstAddr:          "fd00:0:0:1::5",
				topologyFilePath: testFileDir + "test1.txt",
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "no key that matches source address",
			args: args{
				srcAddr:          []string{"fd00:0:0:2::1"},
				dstAddr:          "fd00:0:0:1::5",
				topologyFilePath: testFileDir + "test1.txt",
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "no key that matches destination address",
			args: args{
				srcAddr:          []string{"fd00:0:0:1::1"},
				dstAddr:          "fd00:0:0:2::5",
				topologyFilePath: testFileDir + "test1.txt",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//graph, err := dijkstra.Import(tt.args.topologyFilePath)
			//if err != nil {
			//	t.Errorf("failed to import graph from file : %v", err)
			//	return
			//}
			graph := makeTestGraph()
			got, err := MakeSIDList(graph, tt.args.srcAddr, tt.args.dstAddr)
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

func makeTestGraph() *dijkstra.Graph {
	graph := dijkstra.NewGraph()
	graph.AddMappedVertex("fd00:0:0:1::1")
	graph.AddMappedVertex("fd00:0:0:2::1")
	graph.AddMappedVertex("fd00:0:0:3::1")
	graph.AddMappedVertex("fd00:0:0:4::1")
	graph.AddMappedVertex("fd00:0:0:5::1")

	_ = graph.AddMappedArc("fd00:0:0:1::1", "fd00:0:0:2::1", 4)
	_ = graph.AddMappedArc("fd00:0:0:1::1", "fd00:0:0:3::1", 2)

	_ = graph.AddMappedArc("fd00:0:0:2::1", "fd00:0:0:4::1", 2)
	_ = graph.AddMappedArc("fd00:0:0:2::1", "fd00:0:0:3::1", 3)
	_ = graph.AddMappedArc("fd00:0:0:2::1", "fd00:0:0:5::1", 3)

	_ = graph.AddMappedArc("fd00:0:0:3::1", "fd00:0:0:2::1", 1)
	_ = graph.AddMappedArc("fd00:0:0:3::1", "fd00:0:0:4::1", 4)
	_ = graph.AddMappedArc("fd00:0:0:3::1", "fd00:0:0:5::1", 5)
	return graph
}

// TODO more cases
func Test_makeGraph(t *testing.T) {
	g, _ := dijkstra.Import(testFileDir + "test2.txt")
	type args struct {
		nodesInfo *api.NodesInfo
	}
	tests := []struct {
		name    string
		args    args
		want    *dijkstra.Graph
		wantErr bool
	}{
		{
			name: "correct case",
			args: args{
				nodesInfo: &api.NodesInfo{
					Nodes: []*api.Node{
						{
							SID:            "fd00:0:0:1::1",
							NextSids:       []string{"fd00:0:0:2::1", "fd00:0:0:3::1"},
							LinkUsageRatio: 4,
						}, {
							SID:            "fd00:0:0:2::1",
							NextSids:       []string{"fd00:0:0:4::1"},
							LinkUsageRatio: 4,
						}, {
							SID:            "fd00:0:0:3::1",
							NextSids:       []string{"fd00:0:0:5::1"},
							LinkUsageRatio: 1,
						}, {
							SID:            "fd00:0:0:4::1",
							NextSids:       []string{"fd00:0:0:5::2", "fd00:0:0:6::1"},
							LinkUsageRatio: 3,
						}, {
							SID:            "fd00:0:0:5::1",
							NextSids:       []string{"fd00:0:0:4::2", "fd00:0:0:6::1"},
							LinkUsageRatio: 2,
						}, {
							SID: "fd00:0:0:6::1",
						},
					},
				},
			},
			want:    &g,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeGraph(tt.args.nodesInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeGraph() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				if tt.want == nil {
					return
				} else if tt.want != nil {
					t.Errorf("makeGraph() got = %v, want %v", got, tt.want)
				}
			}

			//if !reflect.DeepEqual(got.Verticies, tt.want.Verticies) {
			//	t.Errorf("makeGraph() got = %v, want %v", got.Verticies, tt.want.Verticies)
			//}
		})
	}
}
