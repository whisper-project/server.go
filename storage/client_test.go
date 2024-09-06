package storage

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestClientData_HasChanged(t *testing.T) {
	ctx := context.Background()
	received := ClientData{Id: uuid.New().String(), ProfileId: uuid.New().String(), IsPresenceLogging: 1}
	yes, why := received.HasChanged(ctx)
	if yes != true || why != "APNS token from new" {
		t.Errorf("HasChanged() = %v, %q; want %v, %q", yes, why, true, "APNS token from new")
	}
	if err := SaveFields(ctx, &received); err != nil {
		t.Errorf("Failed to save stored data for client %q: %v", received.Id, err)
	}
	yes, why = received.HasChanged(ctx)
	if yes != false || why != "" {
		t.Errorf("HasChanged() = %v, %q; want %v, %q", yes, why, false, "")
	}
	received.LastSecret = "secret"
	yes, why = received.HasChanged(ctx)
	if yes != true || why != "unconfirmed secret from existing" {
		t.Errorf("HasChanged() = %v, %q; want %v, %q", yes, why, true, "unconfirmed secret from existing")
	}
	if err := SaveFields(ctx, &received); err != nil {
		t.Errorf("Failed to save stored data for client %q: %v", received.Id, err)
	}
	received.Token = "token"
	yes, why = received.HasChanged(ctx)
	if yes != true || why != "new APNS token from existing" {
		t.Errorf("HasChanged() = %v, %q; want %v, %q", yes, why, true, "new APNS token from existing")
	}
	if err := SaveFields(ctx, &received); err != nil {
		t.Errorf("Failed to save stored data for client %q: %v", received.Id, err)
	}
	received.AppInfo = "app info"
	yes, why = received.HasChanged(ctx)
	if yes != true || why != "new build data from existing" {
		t.Errorf("HasChanged() = %v, %q; want %v, %q", yes, why, true, "new build data from existing")
	}
	if err := SaveFields(ctx, &received); err != nil {
		t.Errorf("Failed to save stored data for client %q: %v", received.Id, err)
	}
	received.IsPresenceLogging = 0
	yes, why = received.HasChanged(ctx)
	if yes != true || why != "no presence logging from existing" {
		t.Errorf("HasChanged() = %v, %q; want %v, %q", yes, why, true, "no presence logging from existing")
	}
	if err := DeleteStorage(ctx, &received); err != nil {
		t.Errorf("Failed to delete stored data for client %q: %v", received.Id, err)
	}
}

func TestCountLegacyClients(t *testing.T) {
	if os.Getenv("DO_LEGACY_TESTS") != "YES" {
		t.Skip("Skipping legacy client test")
	}
	if err := PushConfig("../.env.production"); err != nil {
		t.Fatalf("Can't load production config: %v", err)
	}
	defer PopConfig()
	data := ClientData{}
	count := 0
	countOnly := func() {
		count++
	}
	ctx := context.Background()
	if err := MapFields(ctx, countOnly, &data); err != nil {
		if errors.Is(err, ctx.Err()) {
			t.Logf("Found %d production clients before timing out.", count)
		} else {
			t.Errorf("Failed to map production data: %v", err)
		}
	} else {
		t.Logf("Found %d clients in production.", count)
	}
}
