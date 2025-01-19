package com.qapsoft.io

import java.nio.charset.Charset
import java.nio.charset.StandardCharsets

fun BinaryTableData.indexOf(value:Int,columnIndex:Int, start:Int=0):Int{
    for (i in start until header.maxRowsCount ){
        if(getInt(i, columnIndex)==value)
            return i
    }
    return -1
}
fun BinaryTableData.indexOf(value:Long,columnIndex:Int, start:Int=0):Int{
    for (i in start until header.maxRowsCount ){
        if(getLong(i, columnIndex)==value)
            return i
    }
    return -1
}
fun BinaryTableData.indexOf(value:Double,columnIndex:Int, start:Int=0):Int{
    for (i in start until header.maxRowsCount ){
        if(getDouble(i, columnIndex)==value)
            return i
    }
    return -1
}
fun BinaryTableData.indexOf(value:String,columnIndex:Int, start:Int=0, charset: Charset=StandardCharsets.UTF_8):Int{
    for (i in start until header.maxRowsCount ){
        if(getString(i, columnIndex,charset)==value)
            return i
    }
    return -1
}

fun BinaryTableData.indexOf(value:Int,columnName:String, start:Int=0):Int{
    for (i in start until header.maxRowsCount ){
        if(getInt(i, columnName)==value)
            return i
    }
    return -1
}
fun BinaryTableData.indexOf(value:Long,columnName:String, start:Int=0):Int{
    for (i in start until header.maxRowsCount ){
        if(getLong(i, columnName)==value)
            return i
    }
    return -1
}
fun BinaryTableData.indexOf(value:Double,columnName:String, start:Int=0):Int{
    for (i in start until header.maxRowsCount ){
        if(getDouble(i, columnName)==value)
            return i
    }
    return -1
}
fun BinaryTableData.indexOf(value:String,columnName:String, start:Int=0, charset: Charset=StandardCharsets.UTF_8):Int{
    for (i in start until header.maxRowsCount ){
        if(getString(i, columnName,charset)==value)
            return i
    }
    return -1
}