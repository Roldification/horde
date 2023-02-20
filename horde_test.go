package horde

import (
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	dbx "github.com/go-ozzo/ozzo-dbx"
)

type Customer struct {
	CustomerID string `db:"CustomerID"`
	LastName   string `db:"LastName"`
	FirstName  string `db:"FirstName"`
}

func (c Customer) TableName() string {
	return "Customer"
}

func (c Customer) SavingsAccounts() Relationship {
	return Relationship{
		PrimaryKey:   "CustomerID",
		ForeignKey:   "FKCustomerIDAccount",
		RelateModel:  SavingsAccount{},
		Relationship: HasMany,
	}
}

type SavingsAccount struct {
	AccountNumber        string  `db:"AccountNumber"`
	FKCustomerIDAccount  string  `db:"FKCustomerIDAccount"`
	FKSAProductIDAccount string  `db:"FKSAProductIDAccount"`
	Balance              float64 `db:"Balance"`
}

func (sa SavingsAccount) TableName() string {
	return "SavingsAccount"
}

func TestHordeFindOne(t *testing.T) {
	mockDb, err := dbx.Open("mssql", "odbc:server=.\\mssql;port=1433;user id=sa;database=ICFS_PAGADIAN_06302022;password={r8d/1ct5041};")

	if err != nil {
		t.Error(err)
	}

	customer := Model{
		BaseTable: Customer{},
	}

	myRes, err2 := customer.FindOne().Where(WhereClause{
		Column:    "CustomerID",
		Value:     "026-ABC",
		Condition: "=",
	}).Get(mockDb)

	_, ok := myRes.(map[string]interface{})

	if !ok {
		t.Error("Must return a map of string to interface")
	}

	if err2 != nil {
		t.Error(err2.Error())
	}
}

func TestHordeRelationshipHasMany(t *testing.T) {
	mockDb, err := dbx.Open("mssql", "odbc:server=.\\mssql;port=1433;user id=sa;database=ICFS_PAGADIAN_06302022;password={r8d/1ct5041};")

	if err != nil {
		t.Error(err)
	}

	customer := Model{
		BaseTable: Customer{},
	}

	myRes, err2 := customer.FindOne().Join("SavingsAccounts", []WhereClause{}).Where(WhereClause{
		Column:    "CustomerID",
		Value:     "026-0000002",
		Condition: "=",
	}).Get(mockDb)

	value, ok := myRes.(map[string]interface{})

	if !ok {
		t.Error("Must return a map of string to interface")
	}

	_, ok = value["SavingsAccounts"].([]map[string]interface{})

	if !ok {
		t.Error("Must return a relationship for SavingsAccounts")
	}

	if err2 != nil {
		t.Error(err2.Error())
	}
}
