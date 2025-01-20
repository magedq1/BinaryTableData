package com.qapsoft.io

import java.io.ByteArrayInputStream
import java.io.InputStreamReader
import java.nio.charset.StandardCharsets

data class BinaryTableHeader(
    val maxRowsCount:Int,
    val columns:List<BinaryColumn>,
){
    val schemaSize:Int=columns.encodedAsByteArray().size+8 // 2 int schemaBytesSize, maxRowsCount
    val maxRowLength:Int
    val columnsInfoByIndex:Map<Int, ColumnInfo>
    val columnsInfoByName:Map<String, ColumnInfo>

    init {
        columnsInfoByIndex = mutableMapOf<Int, ColumnInfo>().also { mapByIndex->
            var pos = 0
            columnsInfoByName= mutableMapOf()
            columns.forEachIndexed { index, binaryColumn ->
                mapByIndex[index] = ColumnInfo(
                    index,
                    binaryColumn,
                    pos
                )
                columnsInfoByName[binaryColumn.name]=mapByIndex[index]!!
                pos+=binaryColumn.length
            }
            maxRowLength=pos
        }
    }

    data class ColumnInfo(
        val index:Int,
        val binaryColumn: BinaryColumn,
        val startPos:Int
    )
}

