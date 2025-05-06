package repository

import (
	"sync"
)

type BuddyRepository struct {
}

type Buddy struct {
	ID      string
	OwnerID string
	BuddyID string
}

type BuddyDetails struct {
	BuddyEmail string
	Buddy
}

var buddyRepository *BuddyRepository
var buddyRepositoryOnce sync.Once

func GetBuddyRepository() *BuddyRepository {
	buddyRepositoryOnce.Do(func() {
		buddyRepository = &BuddyRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS buddies (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"owner_id uuid NOT NULL, " +
			"buddy_id uuid NOT NULL, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_buddies_owner_id ON buddies(owner_id)")
		if err != nil {
			panic(err)
		}
	})
	return buddyRepository
}

func (r *BuddyRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// No updates yet
}

func (r *BuddyRepository) Create(e *Buddy) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO buddies "+
		"(owner_id, buddy_id) "+
		"VALUES ($1, $2) "+
		"RETURNING id",
		e.OwnerID, e.BuddyID).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *BuddyRepository) GetOne(id string) (*BuddyDetails, error) {
	e := &BuddyDetails{}
	err := GetDatabase().DB().QueryRow("SELECT buddies.id, buddies.owner_id, buddies.buddy_id, "+
		"users.email "+
		"FROM buddies "+
		"INNER JOIN users ON buddies.buddy_id = users.id "+
		"WHERE buddies.id = $1",
		id).Scan(&e.ID, &e.OwnerID, &e.BuddyID, &e.BuddyEmail)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *BuddyRepository) GetAllByOwner(ownerID string) ([]*BuddyDetails, error) {
	var result []*BuddyDetails
	rows, err := GetDatabase().DB().Query("SELECT buddies.id, buddies.owner_id, buddies.buddy_id, "+
		"users.email "+
		"FROM buddies "+
		"INNER JOIN users ON buddies.buddy_id = users.id "+
		"WHERE owner_id = $1 "+
		"ORDER BY id DESC", ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &BuddyDetails{}
		err = rows.Scan(&e.ID, &e.OwnerID, &e.BuddyID, &e.BuddyEmail)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *BuddyRepository) GetMutualBuddies(userID string, buddyIDs []string) ([]*Buddy, error) {
    // Build query placeholders and arguments manually
    query := "SELECT b.id, b.owner_id, b.buddy_id FROM buddies b WHERE b.owner_id IN ("
    args := make([]interface{}, 0, len(buddyIDs)+1)

    for i, id := range buddyIDs {
        if i > 0 {
            query += ", "
        }
        query += "?"
        args = append(args, id)
    }
    query += ") AND b.buddy_id = ?"
    args = append(args, userID)

    rows, err := GetDatabase().DB().Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var mutualBuddies []*Buddy
    for rows.Next() {
        buddy := &Buddy{}
        if err := rows.Scan(&buddy.ID, &buddy.OwnerID, &buddy.BuddyID); err != nil {
            return nil, err
        }
        mutualBuddies = append(mutualBuddies, buddy)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return mutualBuddies, nil
}

func (r *BuddyRepository) AreMutualBuddies(userID string, buddyID string) (bool, error) {
    query := `
        SELECT EXISTS (
            SELECT 1
            FROM buddies b1
            INNER JOIN buddies b2
            ON b1.owner_id = b2.buddy_id AND b1.buddy_id = b2.owner_id
            WHERE b1.owner_id = $1 AND b1.buddy_id = $2
        )
    `
    var exists bool
    err := GetDatabase().DB().QueryRow(query, userID, buddyID).Scan(&exists)
    if err != nil {
        return false, err
    }
    return exists, nil
}

func (r *BuddyRepository) Delete(e *BuddyDetails) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM buddies WHERE id = $1", e.ID)
	return err
}
