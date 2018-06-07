package database

import (
	"os"
	"testing"
	"time"
	//node "skywire_uptime/node"
)

var db *DbConn
var dbName = "testing.db"

func TestGetBadUser(t *testing.T) {
	_, err := db.GetNodeByKey("bad")
	if err != nil {
		t.Error("Failed to fail get non-existant user")
	}
}

func TestInsertAndGetUser(t *testing.T) {
	pubKey := "demoKey"
	timeFirst := time.Now()
	timeLast := time.Now()
	timeSeen := int64(1)

	db.InsertNode(pubKey, timeFirst, timeLast, timeSeen)

	user, err := db.GetNodeByKey(pubKey)
	if err != nil {
		t.Error("Failed to get get user")
	}

	if user.PublicKey != pubKey {
		t.Error("Node public key wrong")
	}
	if user.FirstSeen != timeFirst {
		t.Error("Node creation date mismatch")
	}
	if user.LastSeen != timeLast {
		t.Error("Node last seen date mismatch")
	}
	if user.TimesSeen != timeSeen {
		t.Error("Node times seen mismatch")
	}
}

func TestNodeGet(t *testing.T) {
	//ti := time.Now()
	uid := db.InsertNode("demo2", time.Now(), time.Now(), 1)

	user, err := db.GetNodeByKey("demo2")
	if err != nil {
		t.Error("failed to get node by key")
	}

	if uid.PublicKey != user.PublicKey {
		t.Error("PublicKey mismatch in node get")
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
	defer os.Remove(dbName)

	// run the tests
	os.Exit(m.Run())
}
