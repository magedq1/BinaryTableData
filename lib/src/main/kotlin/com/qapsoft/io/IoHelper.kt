package com.qapsoft.io

import java.io.ByteArrayInputStream
import java.io.InputStreamReader
import java.nio.ByteBuffer
import java.nio.charset.Charset
import java.nio.charset.StandardCharsets

fun BinaryTableData.setValue(rowIndex:Int, columnIndex:Int , value:Int){
    setValue(rowIndex, columnIndex, value.toByteArray())
}
fun BinaryTableData.setValue(rowIndex:Int, columnName:String , value:Int){
    setValue(rowIndex, columnName, value.toByteArray())
}

fun BinaryTableData.setValue(rowIndex:Int, columnIndex:Int , value:Long){
    setValue(rowIndex, columnIndex, value.toByteArray())
}
fun BinaryTableData.setValue(rowIndex:Int, columnName:String , value:Long){
    setValue(rowIndex, columnName, value.toByteArray())
}

fun BinaryTableData.setValue(rowIndex:Int, columnIndex:Int , value:Double){
    setValue(rowIndex, columnIndex, value.toByteArray())
}
fun BinaryTableData.setValue(rowIndex:Int, columnName:String , value:Double){
    setValue(rowIndex, columnName, value.toByteArray())
}

fun BinaryTableData.setValue(rowIndex:Int, columnIndex:Int , value:String, charset: Charset=StandardCharsets.UTF_8){
    setValue(rowIndex, columnIndex, value.toByteArray(charset))
}
fun BinaryTableData.setValue(rowIndex:Int, columnName:String , value:String, charset: Charset=StandardCharsets.UTF_8){
    setValue(rowIndex, columnName, value.toByteArray(charset))
}

fun BinaryTableData.getInt(rowIndex:Int, columnIndex:Int):Int{
    val bytes = getValue(rowIndex, columnIndex)
    return bytes.asInt()
}

fun BinaryTableData.getInt(rowIndex:Int, columnName:String):Int{
    val bytes = getValue(rowIndex, columnName)
    return bytes.asInt()
}

fun BinaryTableData.getLong(rowIndex:Int, columnIndex:Int):Long{
    val bytes = getValue(rowIndex, columnIndex)
    return bytes.asLong()
}

fun BinaryTableData.getLong(rowIndex:Int, columnName:String):Long{
    val bytes = getValue(rowIndex, columnName)
    return bytes.asLong()
}

fun BinaryTableData.getDouble(rowIndex:Int, columnIndex:Int):Double{
    val bytes = getValue(rowIndex, columnIndex)
    return bytes.asDouble()
}

fun BinaryTableData.getDouble(rowIndex:Int, columnName:String):Double{
    val bytes = getValue(rowIndex, columnName)
    return bytes.asDouble()
}

fun BinaryTableData.getString(rowIndex:Int, columnIndex:Int, charset: Charset=StandardCharsets.UTF_8):String{
    val bytes = getValue(rowIndex, columnIndex)
        .filter { it != 0.toByte() }.toByteArray()
    return String(bytes, charset)
}

fun BinaryTableData.getString(rowIndex:Int, columnName:String, charset: Charset=StandardCharsets.UTF_8):String{
    val bytes = getValue(rowIndex, columnName)
        .filter { it != 0.toByte() }.toByteArray()
    return String(bytes, charset)
}


fun BinaryStreamReader.getBytesAt(pos:Long, offset:Int):ByteArray{
    val res = ByteArray(offset)
    readAt(pos, res, 0, offset)
    return res
}

fun ByteArrayStreamReader.toByteArray():ByteArray{
    return getBytesAt(0, length().toInt())
}
fun ByteArray.asInt():Int{
    val bytes = this
    return (bytes[0].toInt() and 0xFF shl 24) or
            (bytes[1].toInt() and 0xFF shl 16) or
            (bytes[2].toInt() and 0xFF shl 8) or
            (bytes[3].toInt() and 0xFF)
}

fun ByteArray.asDouble():Double{
    val bytes = this
    val buffer = ByteBuffer.wrap(bytes)
    return buffer.double
}

fun ByteArray.asLong():Long{
    val bytes = this
    return (bytes[0].toLong() and 0xFF shl 56) or
            (bytes[1].toLong() and 0xFF shl 48) or
            (bytes[2].toLong() and 0xFF shl 40) or
            (bytes[3].toLong() and 0xFF shl 32) or
            (bytes[4].toLong() and 0xFF shl 24) or
            (bytes[5].toLong() and 0xFF shl 16) or
            (bytes[6].toLong() and 0xFF shl 8) or
            (bytes[7].toLong() and 0xFF)
}

fun BinaryStreamWriter.putLong(pos: Long, value:Long){
    writeAt(pos, value.toByteArray())
}


fun BinaryStreamWriter.putInt(pos: Long, value:Int){
    writeAt(pos, value.toByteArray())
}




fun Int.toByteArray():ByteArray{
    return byteArrayOf(
        (this shr 24).toByte(),
        (this shr 16).toByte(),
        (this shr 8).toByte(),
        this.toByte()
    )
}

fun Double.toByteArray():ByteArray{
    val buffer = ByteBuffer.allocate(Double.SIZE_BYTES)
    buffer.putDouble(this)
    return buffer.array()
}

fun Long.toByteArray():ByteArray{
    val value=this
    return byteArrayOf(
        (value shr 56).toByte(),
        (value shr 48).toByte(),
        (value shr 40).toByte(),
        (value shr 32).toByte(),
        (value shr 24).toByte(),
        (value shr 16).toByte(),
        (value shr 8).toByte(),
        value.toByte()
    )
}

fun List<BinaryColumn>.encodedAsByteArray():ByteArray{
    val stringBuilder = StringBuilder()
    for (col in this){
        stringBuilder
            .append(col.name)
            .append("/")
            .append(col.length)
            .append("/")
            .append(col.columnType.name)
            .appendLine()
    }
    val rawColumns = stringBuilder.toString().toByteArray(StandardCharsets.UTF_8)
    return rawColumns
}

fun ByteArray.asBinaryColumnList():List<BinaryColumn>{
    val result = mutableListOf<BinaryColumn>()
    val reader = InputStreamReader(ByteArrayInputStream(this), StandardCharsets.UTF_8)
    val lines = reader.readLines()
    for(l in lines){
        val colData = l.split("/")
        result.add(
            BinaryColumn(
                name = colData[0],
                length = colData[1].toInt(),
                columnType = BinaryColumn.ColumnType.valueOf(colData[2])
            )
        )
    }
    return result
}