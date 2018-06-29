package revrouting

import "testing"

func TestInit(t *testing.T) {
	if WebhooksRouter.CommandsCount() == 0 {
		t.Fatal("WebhooksRouter.CommandsCount() == 0")
	}
}
