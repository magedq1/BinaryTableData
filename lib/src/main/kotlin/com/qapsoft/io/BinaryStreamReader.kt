package com.qapsoft.io

interface BinaryStreamReader {
    fun readAt(pos:Long,buffer:ByteArray, start:Int, offset:Int):Int
    fun readAt(pos:Long,buffer:ByteArray):Int
    fun length():Long
}