package dpoolcsv_test

import (
	"fmt"
	"testing"

	"github.com/fangpinsern/dpoolcsv-go"
	"github.com/stretchr/testify/assert"
)

type User struct {
	FirstName string `dpool:"firstname"`
	LastName  string `dpool:"lastname"`
	Age       int64  `dpool:"age"`
	UserId    string `dpool:"userid"`
}

type Food struct {
	Name  string `dpool:"name"`
	Price int64  `dpool:"price"`
}

func TestDpoolCsv(t *testing.T) {

	t.Parallel()

	newDpoolStore := dpoolcsv.NewDB()

	newDpoolStore.Ingest("/data")

	getUser := &User{}

	err := newDpoolStore.Get(getUser, 0) // get first user

	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, "john", getUser.FirstName)
	assert.Equal(t, "doe", getUser.LastName)
	assert.Equal(t, "1", getUser.UserId)
	assert.Equal(t, int64(21), getUser.Age)
}

func TestDpoolCsvFilter(t *testing.T) {
	t.Parallel()

	newDpoolStore := dpoolcsv.NewDB()

	newDpoolStore.Ingest("/data")

	getUsers := make([]*User, 0)

	err := newDpoolStore.Filter(&getUsers, "firstname", func(columnVal string) bool {
		return columnVal == "john"
	})
	fmt.Println(err)

	assert.Equal(t, 1, len(getUsers))
	assert.Equal(t, "john", getUsers[0].FirstName)
	assert.Equal(t, "doe", getUsers[0].LastName)

	getFoods := make([]*Food, 0)
	limit := int64(20)

	err = newDpoolStore.Filter(&getFoods, "price", func(columnValue int64) bool {
		return columnValue <= limit
	})
	assert.Nil(t, err)

	assert.Equal(t, 2, len(getFoods))
}

func TestDpoolCsvSet(t *testing.T) {
	t.Parallel()

	newDpoolStore := dpoolcsv.NewDB()

	newDpoolStore.Ingest("/data")

	getUsers := make([]*User, 0)
	firstName := "Fang"
	lastName := "PS"
	age := int64(21)
	userId := "4"

	newSetUser := &User{
		FirstName: firstName,
		LastName:  lastName,
		Age:       age,
		UserId:    userId,
	}

	err := newDpoolStore.Set(newSetUser)
	if err != nil {
		fmt.Println(err)
	}

	err = newDpoolStore.Filter(&getUsers, "firstname", func(columnVal string) bool {
		return columnVal == "Fang"
	})

	fmt.Println(err)

	assert.Equal(t, 1, len(getUsers))
	assert.Equal(t, firstName, getUsers[0].FirstName)
	assert.Equal(t, lastName, getUsers[0].LastName)

	getFoods := make([]*Food, 0)
	limit := int64(20)

	err = newDpoolStore.Filter(&getFoods, "price", func(columnValue int64) bool {
		return columnValue <= limit
	})
	assert.Nil(t, err)

	assert.Equal(t, 2, len(getFoods))
}
