package com.qapsoft.io

interface BinaryStreamWriter {
    fun writeAt(pos:Long, buffer:ByteArray, start:Int, offset:Int)
    fun writeAt(pos:Long,buffer:ByteArray)
}