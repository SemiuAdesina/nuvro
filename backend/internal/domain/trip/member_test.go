package trip_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
)

func TestRoles(t *testing.T) {
	for _, test := range []struct {
		role               trip.Role
		read, edit, manage bool
	}{
		{trip.Owner, true, true, true}, {trip.Editor, true, true, false}, {trip.Viewer, true, false, false}, {"unknown", false, false, false},
	} {
		if test.role.CanRead() != test.read || test.role.CanEdit() != test.edit || test.role.CanManageMembers() != test.manage {
			t.Fatalf("role %s", test.role)
		}
	}
}
