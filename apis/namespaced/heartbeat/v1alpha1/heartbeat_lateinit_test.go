// SPDX-FileCopyrightText: 2026 Nik Ogura <nik.ogura@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLateInitializeSkipsPingState guards against a Heartbeat that updates
// itself on every reconcile.
//
// Rootly moves last_pinged_at and expires_at forward on every ping and
// derives status from them. Copied into the spec, they go stale with the next
// ping, and the provider would send the old values back on every reconcile.
// None of them may be late-initialized; they stay in status.atProvider.
func TestLateInitializeSkipsPingState(t *testing.T) {
	t.Parallel()

	name := "nightly-backup"
	tr := &Heartbeat{}
	tr.Spec.ForProvider.Name = &name

	observed := []byte(`{
		"name": "nightly-backup",
		"interval_unit": "minutes",
		"status": "active",
		"last_pinged_at": "2026-01-01T00:00:00Z",
		"expires_at": "2026-01-01T00:10:00Z"
	}`)

	changed, err := tr.LateInitialize(observed)
	require.NoError(t, err)

	assert.True(t, changed, "fields Rootly does not own are still late-initialized")
	require.NotNil(t, tr.Spec.ForProvider.IntervalUnit)
	assert.Equal(t, "minutes", *tr.Spec.ForProvider.IntervalUnit)

	assert.Nil(t, tr.Spec.ForProvider.LastPingedAt, "last_pinged_at must not be late-initialized")
	assert.Nil(t, tr.Spec.ForProvider.ExpiresAt, "expires_at must not be late-initialized")
	assert.Nil(t, tr.Spec.ForProvider.Status, "status must not be late-initialized")
}
