package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInsertAndRetrieve(t *testing.T) {
	conn := GetGorm(GetTestMysSQLConnStr())

	maps1 := []string{"Battle for Itza", "Black Ark", "Arnheim", "Glade of the Everqueen", "Proving Grounds"}
	maps2 := []string{"Battle for Itza", "Black Ark", "Arnheim", "Glade of the Everqueen", "Proving Grounds", "Death's Pass"}

	InsertMapSet(conn, maps1, All)
	InsertMapSet(conn, maps2, All)

	_, mapSet := GetLatestMapSet(conn, All)

	assert.Equal(t, maps2, mapSet.Maps)
}
