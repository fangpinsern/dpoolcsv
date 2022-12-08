# File Structure

The CSV data has to be put in a specific file struture for us to know where the data is stored. Since all CSV data pulled in by Go is in string format, we required some way for us to know what is the columns data type.

## Details

With this in mind, this is the proposed file structure:

```
data
├── food
│   ├── data.csv
│   └── types.csv
└── user
    ├── data.csv
    └── types.csv
```

Inside your data folder will be individual subfolders. Each subfolder will contain the 2 csv files.

| File      | Description                                                                                                                                                  |
| --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| data.csv  | This file contains all the data records                                                                                                                      |
| types.csv | This file contains information about each column's data type. we currently only support int64 and string. If it is neither, we will treat is as string type. |

In the above example file structure, we will have 2 tables. I food table and another user table.

### Example CSV

data.csv

```
userid,firstname,lastname,age
1,john,doe,21
2,hello,world,100
```

types.csv

```
userid,firstname,lastname,age
string,string,string,int64
```
