# In Memory Stucture

An in memory structure is used to store the records for quick access and ease of manipulation. Access to files is slow. Hence on start up, we pull in all the information in to memory and get the required information from there.

The structure chosen for the current iteration, is to use Go Maps and structs.

## Considerations

1. The structure has to be able to retrieve records in constant time if the index is known.
2. Structure chosen needs to support a wide range to data types.

## Details

To retrieve records in constant time, we are going to use Go `maps`. The key in the map is the table name (string) and the value will be an address to the struct named `Table`.

```go
// DB data structure
map[string]*Table
```

### Table struct

The `Table` struct contains all information that concerns the the data in the table. This information includes:

1. Column type
2. Column index
3. Records

```go
type Table struct {
    Records []map[string]interface{}
    ColumnType map[string]reflect.Kind
    ColumnIndex map[string]int
    Writer *csv.Writer
}
```

| Field       | Description                                                                                                                                                                                                                                    |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Records     | List of all records present in the csv file. Each element in the list is a map with the key being the column name. An interface{} is used to ensure that we are able to support multiple data types                                            |
| ColumnType  | The data type of the particular column. All values in the same column should have the same data type. `Reflect.Kind` is an enum that stores the datatype in a form the Go reflect understands.                                                 |
| ColumnIndex | The index which the data is in the record. This is needed as when csv file data are pulled in, it is represented as a slice/list of strings. Hence when we want to write into the csv file, we need to know the which column is at which index |
| Writer      | A writer to the csv is stored allowing us to write to the csv file if there are any updates to records or new records to be added. This writer is opened at data ingest/start time.                                                            |

## Future Feature Updates

Above are things already implemented. The following are features we plan to introduce.

### Indexing

Currently the values are just indexed by the row they are in. Hence you can only reference the records by their row number for constant time retrieval. If you need to find information that fits a sepcific criteria, then you will need to filter through the records to check that specific column to see if it fits the criteria.

An index will take a column (preferbally with unique values) and create a map between the column value and the row it is in. For example, in a Users table, we can index the user by their userID (which should be unique). Hence when we search for the userId, we do not need to look through the whole table to find it.

To support indexing, we can add another field in the table struct. This `Index` field will store information about.

```go
Index map[string]int
```
