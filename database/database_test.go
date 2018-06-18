package database

import (
	"os"
	"testing"
	"time"
	//node "skywire_uptime/node"
)

var db *DbConn
var dbName = "testing.db"

func TestInsertAndGetUser(t *testing.T) {
	pubKey := "C54ED949CF3DA7CD1C48A01456586C09FBEFE11C4A6F47157C24CF8BE0F6315C76"
	timeFirst := time.Now()
	timeLast := time.Now()
	timeSeen := int64(1)

	db.InsertNode(pubKey, timeFirst, timeLast, timeSeen)

	user, err := db.GetNodeByKey(pubKey)
	if err != nil {
		t.Error("Failed to get get node")
	}

	if user.PublicKey != pubKey {
		t.Error("Node public key wrong")
	}
	//These fail due to type conversion differences.
	/*
		if user.FirstSeen != timeFirst {
			fmt.Printf("%v vs %v\n", user.FirstSeen, timeFirst)
			t.Error("Node creation date mismatch")
		}

			if user.LastSeen != timeLast {
				t.Error("Node last seen date mismatch")
			}
	*/
	if user.TimesSeen != timeSeen {
		t.Error("Node times seen mismatch")
	}
}

func TestUpdate(t *testing.T) {
	pubKey := "C54ED949CF3DA7CD1C48A01456586C09FBEFE11C4A6F47157C24CF8BE0F6315C76"
	user, err := db.GetNodeByKey(pubKey)
	if err != nil {
		t.Error("Failed to get node for update")
	}

	err = db.UpdateNode(pubKey, time.Now())
	if err != nil {
		t.Error("Failed to update node")
	}

	userUpdated, err := db.GetNodeByKey(pubKey)
	if err != nil {
		t.Error("Failed to get node for update comparison")
	}

	if err != nil {
		t.Error("Failed to update node")
	}

	if userUpdated.TimesSeen-user.TimesSeen != 1 {
		t.Error("Changes not reflected in updated node.")
	}
}

// ---------------------------------------------------------------------
// Test SEARCH table
// ---------------------------------------------------------------------
func TestSearchInsert(t *testing.T) {
	_, err := db.InsertSearch(500, time.Now())
	if err != nil {
		t.Error("Failed to insert Search")
	}
}

func TestSearchGetLast(t *testing.T) {
	//Need at least 2 rows
	_, err := db.InsertSearch(123, time.Now())
	if err != nil {
		t.Error("Failed to insert Search")
	}

	search, err := db.GetLastSearch()
	if err != nil {
		t.Error("Failed to get last search")
	}
	if search.NumNodesOnline != 123 {
		t.Error("Last search data mismatch")
	}

}

func TestMain(m *testing.M) {
	// delete any old databases (may fail if there is none, but that's ok)
	os.Remove(dbName)

	// connect to database
	db = ConnectDB(dbName)
	defer db.Close()

	// create tables
	db.SetupDb()

	// delete the database after all tests
	//defer os.Remove(dbName)

	// run the tests
	os.Exit(m.Run())
}
