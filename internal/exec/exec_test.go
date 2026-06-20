package exec

import (
	"testing"

	"github.com/Goodmain/cch/internal/config"
)

func TestAssemble(t *testing.T) {
	cmd := config.Command{Command: "php artisan migrate"}
	cases := []struct {
		name string
		in   Inputs
		want string
	}{
		{
			name: "args and params in order",
			in:   Inputs{Args: []string{"prod", "users"}, Params: []string{"--step=3", "--force"}},
			want: "php artisan migrate prod users --step=3 --force",
		},
		{
			name: "skipped optional param omitted",
			in:   Inputs{Args: []string{"prod"}, Params: []string{"--step=3"}},
			want: "php artisan migrate prod --step=3",
		},
		{
			name: "no args no params",
			in:   Inputs{},
			want: "php artisan migrate",
		},
		{
			name: "empty arg skipped",
			in:   Inputs{Args: []string{"prod", ""}},
			want: "php artisan migrate prod",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Assemble(cmd, tc.in); got != tc.want {
				t.Fatalf("Assemble = %q, want %q", got, tc.want)
			}
		})
	}
}
