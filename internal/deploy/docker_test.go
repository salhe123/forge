package deploy

import "testing"

func TestContainerName(t *testing.T) {
	got := ContainerName("Payments API")
	if got != "forge-payments-api" {
		t.Fatalf("got %q", got)
	}
}
