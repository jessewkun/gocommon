package mongodb

import "testing"

func TestBuildMongoURI(t *testing.T) {
	tests := []struct {
		name    string
		uris    []string
		want    string
		wantErr bool
	}{
		{
			name:    "empty",
			uris:    nil,
			wantErr: true,
		},
		{
			name: "single",
			uris: []string{"mongodb://user:pass@host1:27017/db?replicaSet=rs0"},
			want: "mongodb://user:pass@host1:27017/db?replicaSet=rs0",
		},
		{
			name: "multi_with_scheme",
			uris: []string{
				"mongodb://host1:27017/db?replicaSet=rs0",
				"host2:27017",
				"mongodb://host3:27017",
			},
			want: "mongodb://host1:27017,host2:27017,host3:27017/db?replicaSet=rs0",
		},
		{
			name: "multi_no_scheme",
			uris: []string{"host1:27017", "host2:27017"},
			want: "mongodb://host1:27017,host2:27017",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildMongoURI(tt.uris)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
