package com.qapsoft.io

data class BinaryTableHeader(
    val maxRowsCount:Int,
    val columns:List<BinaryColumn>,
    val schemaSize:Int=4096,

){
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