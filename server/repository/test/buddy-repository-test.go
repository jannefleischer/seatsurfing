package repository_test

import (
    "testing"

    "github.com/seatsurfing/seatsurfing/server/repository"
    "github.com/stretchr/testify/assert"
)

func TestAreMutualBuddies(t *testing.T) {
    // Setup: Create a test database and repository
    db := repository.GetDatabase()
    repo := repository.GetBuddyRepository()

    // Clean up the buddies table before starting
    _, err := db.DB().Exec("DELETE FROM buddies")
    assert.NoError(t, err)

    // Insert test data
    userID := "user-1"
    buddyID := "user-2"

    // Create mutual buddy relationship
    _, err = db.DB().Exec("INSERT INTO buddies (owner_id, buddy_id) VALUES ($1, $2)", userID, buddyID)
    assert.NoError(t, err)
    _, err = db.DB().Exec("INSERT INTO buddies (owner_id, buddy_id) VALUES ($1, $2)", buddyID, userID)
    assert.NoError(t, err)

    // Test: Check if they are mutual buddies
    isMutual, err := repo.AreMutualBuddies(userID, buddyID)
    assert.NoError(t, err)
    assert.True(t, isMutual, "Expected user-1 and user-2 to be mutual buddies")

    // Test: Check if a non-mutual relationship is detected
    nonBuddyID := "user-3"
    isMutual, err = repo.AreMutualBuddies(userID, nonBuddyID)
    assert.NoError(t, err)
    assert.False(t, isMutual, "Expected user-1 and user-3 not to be mutual buddies")

    // Clean up after the test
    _, err = db.DB().Exec("DELETE FROM buddies")
    assert.NoError(t, err)
}