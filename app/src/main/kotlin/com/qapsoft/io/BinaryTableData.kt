package com.qapsoft.io

import java.io.File

open class BinaryTableData(open val header: BinaryTableHeader,
                            open val streamReader: BinaryStreamReader?=null,
                            open  val streamWriter: BinaryStreamWriter?=null) {
    constructor(header: BinaryTableHeader,stream: BinaryStream):this(header, stream, stream)
    constructor(header: BinaryTableHeader,file: File):this(header,
        BinaryStreamImpl(file, header.schemaSize+
            header.maxRowsCount*header.maxRowLength
        ))


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

    protected open fun getPos(rowIndex: Int, columnIndex: Int): Long {
        return (header.schemaSize+
                (rowIndex*header.maxRowLength)+
                (header.columnsInfoByIndex[columnIndex]!!.startPos)).toLong()

    }
}