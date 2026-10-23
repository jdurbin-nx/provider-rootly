// SPDX-FileCopyrightText: 2026 Nik Ogura <nik.ogura@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLateInitializeSkipsDerivedComponentFields guards against a
// StatusPageComponent that updates itself on every reconcile.
//
// A catalog-backed component takes its name and description from its source,
// and Rootly computes its status. Copied into the spec, a rename of the
// source or a status change leaves the spec stale, and the provider would
// send the old values back on every reconcile. None of them may be
// late-initialized.
func TestLateInitializeSkipsDerivedComponentFields(t *testing.T) {
	t.Parallel()

	sourceType := "Service"
	tr := &StatusPageComponent{}
	tr.Spec.ForProvider.SourceType = &sourceType

	observed := []byte(`{
		"source_type": "Service",
		"name": "Checkout",
		"description": "Takes payments",
		"position": 1,
		"status": "operational"
	}`)

	changed, err := tr.LateInitialize(observed)
	require.NoError(t, err)

	assert.True(t, changed, "fields Rootly does not derive are still late-initialized")
	require.NotNil(t, tr.Spec.ForProvider.Position)
	assert.InDelta(t, 1.0, *tr.Spec.ForProvider.Position, 0)

	assert.Nil(t, tr.Spec.ForProvider.Name, "name must not be late-initialized")
	assert.Nil(t, tr.Spec.ForProvider.Description, "description must not be late-initialized")
	assert.Nil(t, tr.Spec.ForProvider.Status, "status must not be late-initialized")
}
