package com.qapsoft.io

class ByteArrayStreamReader(val bytes:ByteArray):BinaryStreamReader {
    override fun readAt(pos: Long, buffer: ByteArray, start: Int, offset: Int): Int {
        if(pos>bytes.size)
            throw ArrayIndexOutOfBoundsException()
        var end=start+offset
        if(end>bytes.size)
            end=bytes.size
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