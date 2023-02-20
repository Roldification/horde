package horde

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	dbx "github.com/go-ozzo/ozzo-dbx"
)

type Model struct {
	BaseTable        interface{}
	queryColumns     string
	orderByColumns   string
	whereClauseQuery string
	joinClauseQuery  string
	tableName        string
	conditionValues  []WhereClause
	joinLevels       []JoinLevels
	queryType        string
}

type JoinLevels struct {
	relationshipLink []string
	level            int
	joinName         string
	currentRow       int
}

type WhereClause struct {
	Column    string
	Condition string
	Value     interface{}
	parseTo   string
}

type Relationship struct {
	PrimaryKey   string
	ForeignKey   string
	RelateModel  interface{}
	Relationship string
}

type TranResult struct {
	LastInsertedId int
	RowsAffected   int64
}

const (
	HasMany = "01"

	HasOne = "00"

	BelongsToMany = "02"

	BelongsToOne = "03"
)

func (h *Model) FindAll() *Model {
	h.queryType = "FindAll"
	var columns []string
	var orderByCols []string
	s := reflect.ValueOf(h.BaseTable)

	table := s.MethodByName("TableName").Call([]reflect.Value{})

	h.tableName = table[0].String()

	for i := 0; i < s.NumField(); i++ {
		columns = append(columns, h.tableName+"."+s.Type().Field(i).Tag.Get("db")+" AS "+h.tableName+s.Type().Field(i).Name)
		orderByCols = append(orderByCols, h.tableName+"."+s.Type().Field(i).Tag.Get("db"))
	}

	h.queryColumns = strings.Join(columns, ",")
	h.orderByColumns = strings.Join(orderByCols, ",")
	return h
}

func (h *Model) FindOne() *Model {
	h.queryType = "FindOne"
	var columns []string
	var orderByCols []string
	s := reflect.ValueOf(h.BaseTable)

	table := s.MethodByName("TableName").Call([]reflect.Value{})

	h.tableName = table[0].String()

	for i := 0; i < s.NumField(); i++ {
		columns = append(columns, h.tableName+"."+s.Type().Field(i).Tag.Get("db")+" AS "+h.tableName+s.Type().Field(i).Name)
		orderByCols = append(orderByCols, h.tableName+"."+s.Type().Field(i).Tag.Get("db"))
	}

	h.queryColumns = strings.Join(columns, ",")
	h.orderByColumns = strings.Join(orderByCols, ",")
	return h
}

func (h *Model) Where(clause WhereClause) *Model {
	clause.parseTo = h.tableName + "" + clause.Column
	h.conditionValues = append(h.conditionValues, clause)

	if h.whereClauseQuery == "" {
		h.whereClauseQuery += " WHERE " + h.tableName + "." + clause.Column + " " + clause.Condition + " " + "{:" + h.tableName + "" + clause.Column + "}"
	} else {
		h.whereClauseQuery += " AND " + h.tableName + "." + clause.Column + " " + clause.Condition + " " + "{:" + h.tableName + "" + clause.Column + "}"
	}
	return h
}

func (h *Model) AndWhere(clause WhereClause) *Model {
	clause.parseTo = h.tableName + "" + clause.Column
	h.conditionValues = append(h.conditionValues, clause)
	h.whereClauseQuery += " AND " + h.tableName + "." + clause.Column + " " + clause.Condition + " " + "{:" + h.tableName + "" + clause.Column + "}"
	return h
}

func (h *Model) OrWhere(clause WhereClause) *Model {
	clause.parseTo = h.tableName + "" + clause.Column
	h.conditionValues = append(h.conditionValues, clause)
	h.whereClauseQuery += " OR " + h.tableName + "." + clause.Column + " " + clause.Condition + " " + "{:" + h.tableName + "" + clause.Column + "}"
	return h
}

func (h *Model) Join(model string, whereClauses []WhereClause) *Model {

	modx := strings.Split(model, ".")

	s := reflect.ValueOf(h.BaseTable)

	for ii, v := range modx {
		sM := s.MethodByName(v)
		relationships := sM.Call([]reflect.Value{})
		relationship := relationships[0]
		r := relationship.FieldByName("RelateModel").Elem()

		if ii == len(modx)-1 {
			var columns []string
			var orderByCols []string
			parentTableName := s.MethodByName("TableName").Call([]reflect.Value{})
			relateTables := r.MethodByName("TableName").Call([]reflect.Value{})
			relateTable := relateTables[0]

			for i := 0; i < r.NumField(); i++ {
				columns = append(columns, relateTable.String()+"."+string(r.Type().Field(i).Tag.Get("db")+" AS "+relateTable.String()+r.Type().Field(i).Name))
				orderByCols = append(orderByCols, relateTable.String()+"."+string(r.Type().Field(i).Tag.Get("db")))
			}

			h.queryColumns += "," + strings.Join(columns, ", ")
			h.orderByColumns += "," + strings.Join(orderByCols, ",")
			h.joinClauseQuery += " INNER JOIN " + relateTable.String() + " ON " + parentTableName[0].String() + "." + relationship.FieldByName("PrimaryKey").String() + " = " + relateTable.String() + "." + relationship.FieldByName("ForeignKey").String()

			h.joinLevels = append(h.joinLevels, JoinLevels{
				relationshipLink: modx,
				level:            ii + 1,
				joinName:         v,
			})

			for _, wc := range whereClauses {
				wc.parseTo = relateTable.String() + "" + wc.Column
				h.conditionValues = append(h.conditionValues, wc)
				if h.whereClauseQuery == "" {
					h.whereClauseQuery += " WHERE " + relateTable.String() + "." + wc.Column + " " + wc.Condition + " " + "{:" + relateTable.String() + "" + wc.Column + "}"
				} else {
					h.whereClauseQuery += " AND " + relateTable.String() + "." + wc.Column + " " + wc.Condition + " " + "{:" + relateTable.String() + "" + wc.Column + "}"
				}
			}
		}

		s = r
	}

	return h
}

func (h *Model) Get(dbxx *dbx.DB) (interface{}, error) {

	// fmt.Println("SELECT " + h.queryColumns + " FROM " + h.tableName + " " + h.joinClauseQuery + h.whereClauseQuery + " ORDER BY " + h.orderByColumns)
	q := dbxx.NewQuery("SELECT " + h.queryColumns + " FROM " + h.tableName + " " + h.joinClauseQuery + h.whereClauseQuery + " ORDER BY " + h.orderByColumns)

	for _, v := range h.conditionValues {
		q.Bind(map[string]interface{}{v.parseTo: v.Value})
	}

	if h.queryType == "FindAll" {
		res := []dbx.NullStringMap{}
		err := q.All(&res)
		var finalResult []map[string]interface{}
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		baseModel := reflect.ValueOf(h.BaseTable)
		baseTableName := baseModel.MethodByName("TableName").Call([]reflect.Value{})[0]
		currentRow := 0
		for iii, v := range res {
			finalMap := map[string]interface{}{}
			var emptySlice []map[string]interface{}

			if iii == 0 {
				// getting the row columns of the base table
				for i := 0; i < baseModel.NumField(); i++ {
					// fmt.Println("fieldName: ", baseModel.Type().Field(i).Name, baseModel.Type().Field(i).Type)
					// TODO: check nulls

					// TODO: check types
					finalMap[baseModel.Type().Field(i).Name] = v[baseTableName.String()+baseModel.Type().Field(i).Name].String
				}
				finalResult = append(finalResult, finalMap)

			} else {
				if finalResult[currentRow][baseModel.Type().Field(0).Name] != v[baseTableName.String()+baseModel.Type().Field(0).Name].String {
					currentRow++
					for i := 0; i < baseModel.NumField(); i++ {
						// fmt.Println("fieldName: ", baseModel.Type().Field(i).Name, baseModel.Type().Field(i).Type)
						// TODO: check nulls

						// TODO: check types
						finalMap[baseModel.Type().Field(i).Name] = v[baseTableName.String()+baseModel.Type().Field(i).Name].String
					}
					finalResult = append(finalResult, finalMap)
					for x := 0; x < len(h.joinLevels); x++ {
						h.joinLevels[x].currentRow = 0
					}
				}

			}

			//begin appending the joins.
			for _, vi := range h.joinLevels {
				s := reflect.ValueOf(h.BaseTable)
				tempResult := finalResult[currentRow]
				// drill down the rabbithole.
				for ii, vv := range vi.relationshipLink {
					childMap := map[string]interface{}{}
					sM := s.MethodByName(vv)
					relationship := sM.Call([]reflect.Value{})[0]
					r := relationship.FieldByName("RelateModel").Elem()

					childTableName := r.MethodByName("TableName").Call([]reflect.Value{})[0]

					// check if we have reached the end of the rabbithole.
					if ii == len(vi.relationshipLink)-1 {
						for i := 0; i < r.NumField(); i++ {
							// fmt.Println("type is: ", r.Type().Field(i).Type.Kind() == reflect.String)

							if r.Type().Field(i).Type.Kind() == reflect.String {
								childMap[r.Type().Field(i).Name] = v[childTableName.String()+r.Type().Field(i).Name].String
							} else if r.Type().Field(i).Type.Kind() == reflect.Float64 {

								flVar, _ := strconv.ParseFloat(v[childTableName.String()+r.Type().Field(i).Name].String, 64)

								childMap[r.Type().Field(i).Name] = flVar
							} else if r.Type().Field(i).Type.Kind() == reflect.Int {
								iVal, _ := strconv.ParseInt(v[childTableName.String()+r.Type().Field(i).Name].String, 10, 64)
								childMap[r.Type().Field(i).Name] = iVal
							}
						}

						if _, ok := tempResult[vv]; !ok {
							tempResult[vv] = emptySlice
							tempResult[vv] = append(tempResult[vv].([]map[string]interface{}), childMap)

							h.joinLevels[ii].currentRow++

						} else {
							lastElement := tempResult[vv].([]map[string]interface{})
							lElem := lastElement[len(lastElement)-1]
							if lElem[r.Type().Field(0).Name] != childMap[r.Type().Field(0).Name] {
								tempResult[vv] = append(tempResult[vv].([]map[string]interface{}), childMap)

								h.joinLevels[ii].currentRow++
								for x := ii + 1; x < len(h.joinLevels); x++ {
									h.joinLevels[x].currentRow = 0
								}
							}
						}

					} else {
						defRow := 0

						if h.joinLevels[ii].currentRow > 0 {
							defRow = h.joinLevels[ii].currentRow - 1
						}
						tempResult = tempResult[vv].([]map[string]interface{})[defRow]
					}
					s = r
				}

			}
		}
		return finalResult, nil
	} else if h.queryType == "FindOne" {
		res := []dbx.NullStringMap{}
		err := q.All(&res)
		var finalResult map[string]interface{}
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		baseModel := reflect.ValueOf(h.BaseTable)
		baseTableName := baseModel.MethodByName("TableName").Call([]reflect.Value{})[0]
		for iii, v := range res {
			finalMap := map[string]interface{}{}
			var emptySlice []map[string]interface{}

			if iii == 0 {
				// getting the row columns of the base table
				for i := 0; i < baseModel.NumField(); i++ {
					// TODO: check nulls

					// TODO: check types
					fmt.Println("type is: ", baseModel.Type().Field(i).Type)
					finalMap[baseModel.Type().Field(i).Name] = v[baseTableName.String()+baseModel.Type().Field(i).Name].String
				}
				finalResult = finalMap

			}

			//begin appending the joins.
			for _, vi := range h.joinLevels {
				s := reflect.ValueOf(h.BaseTable)
				tempResult := finalResult
				// drill down the rabbithole.
				for ii, vv := range vi.relationshipLink {
					childMap := map[string]interface{}{}
					sM := s.MethodByName(vv)
					relationship := sM.Call([]reflect.Value{})[0]
					r := relationship.FieldByName("RelateModel").Elem()

					childTableName := r.MethodByName("TableName").Call([]reflect.Value{})[0]

					// check if we have reached the end of the rabbithole.
					if ii == len(vi.relationshipLink)-1 {
						for i := 0; i < r.NumField(); i++ {

							if r.Type().Field(i).Type.Kind() == reflect.String {
								childMap[r.Type().Field(i).Name] = v[childTableName.String()+r.Type().Field(i).Name].String
							} else if r.Type().Field(i).Type.Kind() == reflect.Float64 {

								flVar, _ := strconv.ParseFloat(v[childTableName.String()+r.Type().Field(i).Name].String, 64)

								childMap[r.Type().Field(i).Name] = flVar
							} else if r.Type().Field(i).Type.Kind() == reflect.Int {
								iVal, _ := strconv.ParseInt(v[childTableName.String()+r.Type().Field(i).Name].String, 10, 64)
								childMap[r.Type().Field(i).Name] = iVal
							}
						}

						if _, ok := tempResult[vv]; !ok {
							tempResult[vv] = emptySlice
							tempResult[vv] = append(tempResult[vv].([]map[string]interface{}), childMap)

							h.joinLevels[ii].currentRow++

						} else {
							lastElement := tempResult[vv].([]map[string]interface{})
							lElem := lastElement[len(lastElement)-1]
							if lElem[r.Type().Field(0).Name] != childMap[r.Type().Field(0).Name] {
								tempResult[vv] = append(tempResult[vv].([]map[string]interface{}), childMap)

								h.joinLevels[ii].currentRow++
								for x := ii + 1; x < len(h.joinLevels); x++ {
									h.joinLevels[x].currentRow = 0
								}
							}
						}

					} else {
						defRow := 0

						if h.joinLevels[ii].currentRow > 0 {
							defRow = h.joinLevels[ii].currentRow - 1
						}
						tempResult = tempResult[vv].([]map[string]interface{})[defRow]
					}
					s = r
				}

			}
		}
		return finalResult, nil
	}

	return nil, nil
}

// Inserts or Updates an entity, if a find() method is given.
func (h *Model) Save(props map[string]interface{}, hasIncrement bool, dbxx *dbx.DB) (TranResult, error) {

	t := TranResult{
		LastInsertedId: 0,
		RowsAffected:   0,
	}
	s := reflect.ValueOf(h.BaseTable)
	table := s.MethodByName("TableName").Call([]reflect.Value{})
	tableName := table[0].String()

	// this is insert, because it did not went through a find() method
	if h.tableName == "" {
		selectID := ""

		if hasIncrement {
			selectID = "select ID = convert(bigint, SCOPE_IDENTITY());"
		}

		var keys []string
		var keyReplacements []string
		for k := range props {
			keys = append(keys, k)
			keyReplacements = append(keyReplacements, "{:"+k+"}")
		}
		queryKey := strings.Join(keys, ",")
		queryKeyReplacements := strings.Join(keyReplacements, ",")

		q := dbxx.NewQuery("INSERT INTO " + tableName + " (" + queryKey + ") VALUES (" + queryKeyReplacements + "); " + selectID)

		for k, v := range props {
			q.Bind(map[string]interface{}{k: v})
		}

		if hasIncrement {
			insertedID := dbx.NullStringMap{}
			err := q.One(insertedID)

			if err != nil {
				return t, err
			}

			retNum, err := strconv.ParseInt(insertedID["ID"].String, 10, 64)

			if err != nil {
				return t, err
			}

			t.LastInsertedId = int(retNum)
			if t.LastInsertedId > 0 {
				t.RowsAffected = 1
			}
		} else {
			sqlres, err := q.Execute()

			if err != nil {
				return t, err
			}

			t.RowsAffected, _ = sqlres.RowsAffected()
		}
	} else {
		var keyReplacements []string
		for k := range props {
			keyReplacements = append(keyReplacements, k+" = {:val"+k+"}")
		}

		queryKeyReplacements := strings.Join(keyReplacements, ",")

		q := dbxx.NewQuery("UPDATE " + tableName + " SET " + queryKeyReplacements + " " + h.whereClauseQuery)

		for _, v := range h.conditionValues {
			q.Bind(map[string]interface{}{v.parseTo: v.Value})
		}

		for k, v := range props {
			q.Bind(map[string]interface{}{"val" + k: v})
		}

		sqlres, err := q.Execute()

		if err != nil {
			return t, err
		}

		t.RowsAffected, _ = sqlres.RowsAffected()

	}

	return t, nil
}

// //func (h *HordeModel)
