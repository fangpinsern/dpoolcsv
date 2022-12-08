package dpoolcsv

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

/*
This DB struct will simulate a DB instance.
The key is the table name
The value is a list of maps representing the data in that row of the table.

Currently on supports string and int64 types with no foreign keys
*/
type DB struct {
	Tables map[string]*Table
}

/*
The table struct provides a structure in memory for the tables that will be
in the Database. Each table will have their own table struct isolating the
table information to their own struct.
*/
type Table struct {
	Data        []map[string]interface{}
	Types       map[string]reflect.Kind
	ColumnIndex map[string]int64
	Writer      *csv.Writer
}

/*
Returns a new database instance
*/
func NewDB() *DB {
	return &DB{
		Tables: make(map[string]*Table),
	}
}

/*
Ingest data from a particluar folder
Data csv has to be accompanied by a types file that gives information about
the type of each column

takes in:
folder path where the data is stored

returns:
error if there is any issue with the folder path provided
*/
func (d *DB) Ingest(folderPath string) error {

	cwd, _ := os.Getwd()
	folderLocation := cwd + folderPath
	dirInfo, _ := os.ReadDir(folderLocation)

	for _, tables := range dirInfo {
		tableFolderName := tables.Name()
		if !tables.IsDir() {
			continue
		}
		dirLocation := folderLocation + "/" + tableFolderName

		newTable := &Table{}
		columnNames, columnTypes, tableValues, writer, err := openProcessData(dirLocation)

		if err != nil {
			log.Fatal(err)
		}

		// Process the types
		newTypes := make(map[string]reflect.Kind)
		newColumnIndex := make(map[string]int64)
		for i, v := range columnTypes {
			columnName := columnNames[i]
			newColumnIndex[columnName] = int64(i)
			if v == "int64" {
				newTypes[columnName] = reflect.Int64
			} else {
				newTypes[columnName] = reflect.String
			}
		}

		// process the data
		numRecords := len(tableValues)
		newData := make([]map[string]interface{}, 0)
		for i := 0; i < numRecords; i++ {
			row := tableValues[i]
			newRow := make(map[string]interface{})
			for j, v := range row {
				if columnTypes[j] == "int64" {
					newRow[columnNames[j]], _ = strconv.ParseInt(v, 0, 64)
				} else {
					newRow[columnNames[j]] = v
				}
			}
			newData = append(newData, newRow)
		}

		tableName := strings.Split(tableFolderName, ".")[0]
		newTable.Data = newData
		// d.Data[tableName] = newData
		newTable.Types = newTypes
		// d.Types[tableName] = newTypes
		newTable.ColumnIndex = newColumnIndex
		newTable.Writer = writer
		d.Tables[tableName] = newTable
	}

	return nil
}

/*
Opens the directory and processes the data to write
It also does implicit checks to make sure the data csv
and the types csv all have the same column name in the
same order

returns:
Column name of the tables
Types of each column
Data of the whole table
err - if there is any error
*/
func openProcessData(dirPath string) (columnNames, columnTypes []string, tableData [][]string, writer *csv.Writer, err error) {
	tableInfo, _ := os.ReadDir(dirPath)
	var columnNamesCheck []string

	for _, fi := range tableInfo {
		fileName := fi.Name()
		if fileName == "data.csv" {
			extension := filepath.Ext(fileName)
			if extension != ".csv" {
				fmt.Println("Not csv file")
				continue
			}

			fileInfo, _ := os.OpenFile(dirPath+"/"+fileName, os.O_APPEND|os.O_RDWR, os.ModeAppend)
			reader := csv.NewReader(fileInfo)
			writer = csv.NewWriter(fileInfo)
			records, err := reader.ReadAll()
			if err != nil {
				return nil, nil, nil, nil, err
			}
			fmt.Println(records)
			columnNames = records[0]
			tableData = records[1:]
		} else if fileName == "types.csv" {
			extension := filepath.Ext(fileName)
			if extension != ".csv" {
				fmt.Println("Not csv file")
				continue
			}

			fileInfo, _ := os.Open(dirPath + "/" + fileName)
			reader := csv.NewReader(fileInfo)
			records, err := reader.ReadAll()
			if err != nil {
				return nil, nil, nil, nil, err
			}
			columnNamesCheck = records[0]
			columnTypes = records[1]
		} else {
			err = fmt.Errorf("invalid file names %s", dirPath)
			return
		}
	}

	if len(columnTypes) == 0 {
		err = fmt.Errorf("no column types %s", dirPath)
		return
	}

	if !reflect.DeepEqual(columnNames, columnNamesCheck) {
		err = fmt.Errorf("column names dont match between files %s", dirPath)
		return
	}

	return
}

func (d *DB) CheckData(tableName string) {
	fmt.Println(d.Tables[tableName])
}

/*
Get the tables names that correspond to the struct provided
The struct given by the user will be lowercased. It should
follow the same name as the folder the data is put in. For examples,
if the folder of the data is nameed "user", then it will correspond
to the struct name "User" or "User". This function automatically
lowercase all struct names

If the interface passed in is a slice, we have to call an extra elem
to get the elem type of the list
*/
func getTableName(val interface{}) string {
	valCheckList := reflect.TypeOf(val).Elem()
	var typeName string
	if valCheckList.Kind() == reflect.Slice {
		valType := valCheckList.Elem().Elem()
		typeName = valType.Name()
	} else {
		typeName = valCheckList.Name()
	}
	typeNameSplit := strings.Split(typeName, ".")
	tableName := strings.ToLower(typeNameSplit[len(typeNameSplit)-1])
	return tableName
}

/*
Get gets the data from the database by the index given and writes it
into the address provided. This assumes that the data provided has
the correct types.

We use the struct field `dpool` to map the data in the DB to the
destination struct provided. If the tag does not exist in the DB, we
will not populate that field.
*/
func (d *DB) Get(dst interface{}, index int) error {

	dstType := reflect.TypeOf(dst).Elem()
	dstValue := reflect.ValueOf(dst).Elem()

	tableName := getTableName(dst)

	// table, ok := d.Data[tableName]
	table, ok := d.Tables[tableName]
	if !ok {
		return fmt.Errorf("TABLE DONT EXIST")
	}

	if len(table.Data) <= index {
		return fmt.Errorf("index exceed table size")
	}

	record := table.Data[index]
	for i := 0; i < dstType.NumField(); i++ {
		structField := dstType.Field(i)
		structName := structField.Tag.Get("dpool")

		data, ok := record[structName]
		if !ok {
			fmt.Println("Does not exist")
			continue
		}
		dstValue.Field(i).Set(reflect.ValueOf(data))
	}

	return nil
}

func (d *DB) Filter(dst interface{}, columnName string, filterFunc interface{}) error {
	// dst is of type *[]*interface{}
	// address to a slice storing address to struct

	dstType := reflect.TypeOf(dst).Elem()
	dstValue := reflect.ValueOf(dst).Elem()

	if dstType.Kind() != reflect.Slice {
		log.Println("WARNING: dst is not a slice type")
	}

	dstTypeElemType := dstType.Elem().Elem()

	tableName := getTableName(dst)
	table, ok := d.Tables[tableName]
	if !ok {
		return fmt.Errorf("TABLE DONT EXIST")
	}

	data := table.Data
	types := table.Types

	columnKind, ok := types[columnName]
	if !ok {
		return fmt.Errorf("column name type does not exist")
	}

	isValidFilterFunc := checkFilterFunc(filterFunc, columnKind)

	if !isValidFilterFunc {
		return fmt.Errorf("not valid filterfunc")
	}

	fmt.Println("filterfunc is valid")
	filterFuncValue := reflect.ValueOf(filterFunc)

	for i := 0; i < len(data); i++ {
		record := data[i]
		columnValue := reflect.ValueOf(record[columnName])
		validVal := filterFuncValue.Call([]reflect.Value{columnValue})
		isValid := validVal[0]
		if !isValid.Bool() {
			continue
		}

		newDstElemValue := reflect.New(dstTypeElemType)
		newDstElemType := newDstElemValue.Type().Elem()
		for i := 0; i < newDstElemType.NumField(); i++ {
			structField := newDstElemType.Field(i)
			structName := structField.Tag.Get("dpool")

			data, ok := record[structName]
			if !ok {
				fmt.Println("Record does not exist")
				continue
			}

			newDstElemValue.Elem().Field(i).Set(reflect.ValueOf(data))
		}

		newSlice := reflect.Append(dstValue, newDstElemValue)
		dstValue.Set(newSlice)
	}

	return nil
}

/*
Has to have only input. type of input does not matter but should be returned
Has to out put a bool
*/
func checkFilterFunc(filterFunc interface{}, columnKind reflect.Kind) bool {

	filterFuncType := reflect.TypeOf(filterFunc)
	// filterFuncValue := reflect.TypeOf(filterFunc).In(0).Name()
	// filterFuncValueOut := reflect.TypeOf(filterFunc).Out(0).Name()

	if filterFuncType.Kind() != reflect.Func {
		return false
	}

	if filterFuncType.NumIn() != 1 {
		return false
	}

	if filterFuncType.NumOut() != 1 {
		return false
	}

	filterFuncReturnValue := filterFuncType.Out(0)

	if filterFuncReturnValue.Kind() != reflect.Bool {
		return false
	}

	filterFuncInputValue := filterFuncType.In(0)

	return filterFuncInputValue.Kind() == columnKind
	// if filterFuncInputValue.Kind() != columnKind {
	// 	return false
	// }

	// return true
}

/*
Set added data to the database for that specific table.
The data is added to the table that has the same name as
the src struct provided

If the tables does note exist, an error is returned. Upsert
is not suported
*/
func (d *DB) Set(src interface{}) error {

	tableName := getTableName(src)
	table, ok := d.Tables[tableName]
	if !ok {
		// should it be an upsert?
		return fmt.Errorf("table does not exist for this data")
	}

	srcType := reflect.TypeOf(src).Elem()
	srcValue := reflect.ValueOf(src).Elem()

	newRecord := make(map[string]interface{})

	for i := 0; i < srcType.NumField(); i++ {
		field := srcType.Field(i)
		fieldTag := field.Tag.Get("dpool")

		if field.Type.Kind() == reflect.Int64 {
			newRecord[fieldTag] = int64(srcValue.Field(i).Int())
		} else {
			newRecord[fieldTag] = srcValue.Field(i).String()
		}

	}

	table.Data = append(table.Data, newRecord)

	// set to file
	columnIndex := table.ColumnIndex
	numColumns := len(columnIndex)
	rowRecord := make([]string, numColumns)
	for k, v := range newRecord {
		index := columnIndex[k]
		rowRecord[index] = fmt.Sprintf("%v", v)
	}

	err := appendToCsvFile(table.Writer, rowRecord)
	if err != nil {
		log.Println("Error appending")
		return err
	}

	return nil
}

/*
Appends records to csv file.
Underneath, it uses csv.WriteAll to append into the csv file.
*/
func appendToCsvFile(writer *csv.Writer, record []string) error {
	records := make([][]string, 1)
	records[0] = record
	err := writer.WriteAll(records)

	// err := writer.Write(record)
	// writer.Flush()

	fmt.Println(err)
	return err
}

func batchAppendToCsvFile(writer *csv.Writer, records [][]string) error {

	err := writer.WriteAll(records)
	return err
}

// type User struct {
// 	FirstName string `dpool:"firstname"`
// 	LastName  string `dpool:"lastname"`
// 	Age       int64  `dpool:"age"`
// 	UserId    string `dpool:"userid"`
// }

// type Food struct {
// 	Name   string `dpool:"name"`
// 	Price  int64  `dpool:"price"`
// 	Rating int64  `dpool:"rating"`
// }

// func main() {
// 	dbInstance := NewDB()
// 	dbInstance.Ingest("/data")
// 	dbInstance.CheckData("user")

// 	newUser := &User{
// 		FirstName: "fang",
// 		LastName:  "ps",
// 		Age:       10,
// 		UserId:    "3",
// 	}

// 	dbInstance.Set(newUser)
// 	dbInstance.CheckData("user")

// 	userData := &User{}
// 	dbInstance.Get(userData, 2)

// 	fmt.Println(userData.FirstName)

// 	foodData := &Food{}

// 	dbInstance.Get(foodData, 0)
// 	fmt.Println(foodData.Name)
// }
