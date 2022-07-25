package peering

import (
	"testing"
)

func TestMakePid(t *testing.T) {
	t.Parallel()

	_ = makePid()
}

func TestRandomElement(t *testing.T) {
	t.Parallel()

	maxSize := 0
	_, err := randomElement(maxSize)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	maxSize = 1
	val, err := randomElement(maxSize)
	if err != nil {
		t.Fatal(err)
	}
	if val != 0 {
		t.Fatal("val should be zero")
	}

	num := 1000
	maxSize = 10
	for i := 0; i < num; i++ {
		val, err := randomElement(maxSize)
		if err != nil {
			t.Fatal(err)
		}
		if val < 0 || val >= maxSize {
			t.Fatal("Invalid randomElement call")
		}
	}
}

func Test_makePid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want uint64
	}{
		{
			name: "make PID",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makePid(); got <= tt.want {
				t.Errorf("makePid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_randomElement(t *testing.T) {
	t.Parallel()

	type args struct {
		maxSize int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "Random Element max size 0",
			args:    struct{ maxSize int }{maxSize: 0},
			want:    0,
			wantErr: true,
		},
		{
			name:    "Random Element max size 1",
			args:    struct{ maxSize int }{maxSize: 1},
			want:    0,
			wantErr: false,
		},
		{
			name:    "Random Element max size > 1",
			args:    struct{ maxSize int }{maxSize: 1000},
			want:    -1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := randomElement(tt.args.maxSize)
			if tt.want == -1 {
				if got < 0 {
					t.Errorf("randomElement() got = %v, want %v to be greather than 1", got, tt.want)
				}
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("randomElement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("randomElement() got = %v, want %v", got, tt.want)
			}
		})
	}
}
