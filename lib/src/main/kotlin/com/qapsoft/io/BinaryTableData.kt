package com.qapsoft.io

import java.io.File

open class BinaryTableData(val header: BinaryTableHeader,
                           val streamReader: BinaryStreamReader?=null,
                           val streamWriter: BinaryStreamWriter?=null) {
    constructor(header: BinaryTableHeader,stream: BinaryStream):this(header, stream, stream)
    constructor(header: BinaryTableHeader,file: File):this(header,
        BinaryStreamImplMapped(file, header.schemaSize+
            header.maxRowsCount*header.maxRowLength
        ))

    companion object{
        fun from(file: File):BinaryTableData{
            return from(
                BinaryStreamImplMapped(
                    file = file,
                    size = file.length().toInt()
                )
            )
        }

        fun from(bytes: ByteArray):BinaryTableData{
            return from(ByteArrayStreamReader(bytes))
        }
        fun from(reader: BinaryStreamReader):BinaryTableData{
            val maxRowsCount = reader.getBytesAt(0,4).asInt()
            val schemaBytesSize = reader.getBytesAt(4,4).asInt()
            val columns = reader.getBytesAt(8, schemaBytesSize).asBinaryColumnList()
            return BinaryTableData(
                header = BinaryTableHeader(
                    columns = columns,
                    maxRowsCount = maxRowsCount
                ),
                streamReader=reader
            )
        }
    }
    init {
        val schemaData = header.columns.encodedAsByteArray()
        streamWriter?.writeAt(0, header.maxRowsCount.toByteArray())
        streamWriter?.writeAt(4, schemaData.size.toByteArray())
        streamWriter?.writeAt(8, schemaData)
    }
    fun setValue(rowIndex:Int, columnName:String, bytes:ByteArray){
        setValue(rowIndex, header.columnsInfoByName[columnName]!!.index, bytes)
    }

    open fun setValue(rowIndex:Int, columnIndex:Int, bytes:ByteArray){
        val pos = getPos(rowIndex, columnIndex)
        if(bytes.size > header.columns[columnIndex].length)
            throw OverSizeError()
        streamWriter?.writeAt(pos, bytes)
    }

    fun getValue(rowIndex:Int, columnName: String):ByteArray{
        return getValue(rowIndex, header.columnsInfoByName[columnName]!!.index)
    }

    open fun getValue(rowIndex:Int, columnIndex:Int):ByteArray{
        val pos = getPos(rowIndex, columnIndex)
        val res=ByteArray(header.columns[columnIndex].length)
        streamReader?.readAt(pos, res)
        return res
    }

    open fun getPos(rowIndex: Int, columnIndex: Int): Long {
        return (header.schemaSize+
                (rowIndex*header.maxRowLength)+
                (header.columnsInfoByIndex[columnIndex]!!.startPos)).toLong()

    }
    open fun getAllRows(): ByteArray{
        val start = getPos(0,0)
        val end = streamReader?.length()?:start
        return streamReader?.getBytesAt(start, (end-start).toInt())?:byteArrayOf()
    }
}