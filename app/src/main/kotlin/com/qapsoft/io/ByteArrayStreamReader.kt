package com.qapsoft.io

open class ByteArrayStreamReader(private val bytes:ByteArray):BinaryStreamReader {
    override fun readAt(pos: Long, buffer: ByteArray, start: Int, offset: Int): Int {
        if(pos>length())
            throw ArrayIndexOutOfBoundsException()
        var end=start+offset
        if(end>length())
            end=length().toInt()
        var c=pos.toInt()
        for(i in start until end){
            buffer[i]=bytes[c++]
        }
        return c-pos.toInt()
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        return readAt(pos, buffer, 0, buffer.size)
    }

    override fun length(): Long {
        return bytes.size.toLong()
    }
}