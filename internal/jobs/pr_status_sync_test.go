package jobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPRStatusSyncJob(t *testing.T) {
	job, err := NewPRStatusSyncJob()
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, TypePRStatusSync, job.Type())
}

func TestParsePRStatusSyncPayload(t *testing.T) {
	job, err := NewPRStatusSyncJob()
	require.NoError(t, err)

	payload, err := ParsePRStatusSyncPayload(job)
	require.NoError(t, err)
	assert.NotNil(t, payload)
}