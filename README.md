# Horde


## Summary

- [Description](#description)
- [Installation](#installation)
- [Getting Started](#getting-started)



## Description
Looks like an ORM. Almost an ORM. Except that it's not. Horde is a dead simple Almost-ORM package I wrote for myself and my team at work. This will help us
fetch SQL results just like in a decent ORM way (i.e: nested json result based on the declared relationship, etc).
This uses `ozzo-dbx` for query execution under the hood and for the meantime only works for Microsoft SQL Server >= 2012.
It used ozzo-dbx because I thought ozzo-dbx could provide what we wanted but it cannot, so what I did was I wrote Horde on top of it because I already had it.

I might add support for MySQL later.

## Installation
```
go get github.com/Roldification/horde
```


## Getting Started


Defining the Model

```go
// Customer is a model for Customer table
type Customer struct {
	CustomerID string `db:"CustomerID"`
	LastName   string `db:"LastName"`
	FirstName  string `db:"FirstName"`
}

// TableName returns the name of the table
func (c Customer) TableName() string {
	return "Customer"
}

// Define the Relationship for the Customer model by returning the Relationship struct. The function name is the Relationship Name.
func (c Customer) SavingsAccounts() h.Relationship {
	return h.Relationship{
		PrimaryKey:   "CustomerID",
		ForeignKey:   "FKCustomerIDAccount",
		RelateModel:  SavingsAccount{},
		Relationship: h.HasMany,
	}
}


// Define another Model.
type SavingsAccount struct {
	AccountNumber        string  `db:"AccountNumber"`
	FKCustomerIDAccount  string  `db:"FKCustomerIDAccount"`
	FKSAProductIDAccount string  `db:"FKSAProductIDAccount"`
	Balance              float64 `db:"Balance"`
}

func (sa SavingsAccount) TableName() string {
	return "SavingsAccount"
}
```
- `h.Relationship{}` - the Relationship between the two models, should be returned in defining Horde Relationship.


Usage:

```go
package main

import (
	"fmt"

	h "github.com/Roldification/horde"
	_ "github.com/denisenkom/go-mssqldb"
	dbx "github.com/go-ozzo/ozzo-dbx"
)


func main() {
  
    // setup database connection
    mockDb, err := dbx.Open("mssql", "odbc:server=xxxx;port=1433;user id=xxxx;database=xxxx;password={xxx};")

    if err != nil {
      panic(err)
    }
    
    // initialize the Horde Model
    customer := h.Model{
      BaseTable: Customer{},
    }
    
    // Horde Model has these methods FindOne(), FindAll(), Where(), OrWhere(), AndWhere() and finally, Get()
    result, err := customer.FindOne().Join("SavingsAccounts", []h.WhereClause{}).Where(h.WhereClause{
      Column:    "CustomerID",
      Condition: "=",
      Value:     "026-0000008",
    }).Get(mockDb)

    if err != nil {
      panic(err)
    }
    jsonString, _ := json.Marshal(result)
    fmt.Println(string(jsonString))

}
```
this outputs:
```json
{
  "CustomerID": "026-0000008",
  "FirstName": "JOHN",
  "LastName": "DOE",
  "SavingsAccounts": [
    {
      "AccountNumber": "012345567",
      "Balance": 99999.25,
      "FKCustomerIDAccount": "026-0000008",
      "FKSAProductIDAccount": "01"
    }
  ]
}
```
keywords:
- `h.Model{}` - instantiates our Horde Model. Exposes a BaseTable property to be filled with the Model (struct)
- `FindOne()` - fetches one row (first row) of data from the database
- `FindAll()` - fetches multiple rows from the database
- `Where(), OrWhere(), AndWhere()` - filters the model. This function requires the `WhereClause{}` as parameter.
- `WhereClause{}` - a struct to constructing filter statements in Horde. It has 3 properties: `Column`, `Condition`, and `Value`
- `Join()` - a method to retrieve the relationship data of the parent model. This will require the Relationship Name defined for the model and also a `WhereClause{}`
for filtering the related model.
   The Join() can go deep too. Just use it as `Join("SavingsAccounts.SavingsTransactions.TransactionDetails", []WhereClause{})`
   or you can chain it with multiple joins i.e: `Join(...).Join(...).Get(mockDb)`
- `Get()` - get the data. will return a map[string]interface{} for `FindOne()` and a []map[string]interface{} for `FindAll()`. It will require the ozzo-db connection as parameter.

It has Save() method too for Upsert() operations (like in other ORM). I will document it later ;)
