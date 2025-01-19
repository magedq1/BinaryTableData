# BinaryTableData Project

## Project Description

BinaryTableData is a Kotlin library designed to simplify reading and writing structured binary data in the form of tables. The library provides tools for handling binary columns and binary tables, along with support for data compression and encryption. This library can be used in applications that require efficient storage and processing of binary data.

## Usage

[Instructions on how to build, run, and use the project will be placed here.]


### Getting Started

1. Clone the project.
2. Build the project using Gradle. You can use the command `./gradlew build` in the project root directory.
3. The library can then be included as a dependency in your Kotlin project.

[Further details can be added as needed.]

## Usage Examples

### Creating a Table and Adding Columns

```kotlin
val table = BinaryTableData(
        header = BinaryTableHeader(
            maxRowsCount = 1000,
            columns = listOf(
                BinaryColumn.Int("_id"),
                BinaryColumn.String("first_name"),
                BinaryColumn.String("last_name"),
                BinaryColumn.Raw("raw_data", 8192), 
                )

        ),
        file = File("storedFile")
    )
```

### Writing Data to the Table

```kotlin

import com.qapsoft.io.BinaryColumn
import com.qapsoft.io.BinaryTableData
import com.qapsoft.io.BinaryTableHeader
import com.qapsoft.io.setValue
import java.io.File

fun main() {
	val table = BinaryTableData(/*.....*/)

    //write some data
    val rowId = 0
    table.setValue(rowId, "_id", 1)
    //or you can also use column index instead of column name
    //table.setValue(rowId, 0, 1)

    table.setValue(rowId, "first_name", "Ali")

    table.setValue(rowId, "last_name", "Saleh")

    table.setValue(rowId, "raw_data", ByteArray(8192)) //any raw data


}

```

### Reading Data from the Table

```kotlin
import com.qapsoft.io.BinaryColumn
import com.qapsoft.io.BinaryTableData
import com.qapsoft.io.BinaryTableHeader
import com.qapsoft.io.getInt
import com.qapsoft.io.getString
import java.io.File

fun main() {

    val table = BinaryTableData(/*.....*/)

    //read data

    val rowId = 0
    println(table.getInt(rowId, "_id"))
    println(table.getString(rowId, "first_name"))
    println(table.getString(rowId, "last_name"))

    val rawData: ByteArray = table.getValue(rowId, "raw_data")
}
```

