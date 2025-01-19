package com.qapsoft.io

open class ByteArrayStreamReader(private val bytes:ByteArray):BinaryStreamReader {
    override fun readAt(pos: Long, buffer: ByteArray, start: Int, offset: Int): Int {
        if(pos>length())
            throw ArrayIndexOutOfBoundsException()
        val newOffset = if(pos+offset>length())
                            length()-pos
                        else
                            null

        var end=start + (newOffset?:offset).toInt()
        if(end>buffer.size)
            end=buffer.size
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