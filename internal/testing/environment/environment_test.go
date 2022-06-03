package environment

import "testing"

func TestSetupBadgerDatabase(t *testing.T) {
	t.Parallel()
	db := SetupBadgerDatabase(t)
	if err := db.VerifyChecksum(); err != nil {
		t.Fatal(err)
	}
}
