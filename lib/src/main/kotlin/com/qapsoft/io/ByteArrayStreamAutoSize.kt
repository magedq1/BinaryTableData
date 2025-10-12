package com.qapsoft.io

class ByteArrayStreamAutoSize: ByteArrayStreamWriterAutoSize(),  BinaryStream{
    private var _cacheReader: BinaryStreamReader? = null
    private val reader get() = _cacheReader?:let {
        _cacheReader = ByteArrayStreamReader(toByteArray())
        _cacheReader!!
    }
    override fun readAt(
        pos: Long,
        buffer: ByteArray,
        start: Int,
        offset: Int
    ): Int {
        return reader.readAt(
            pos, buffer, start, offset
        )
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        return reader.readAt(
            pos, buffer
        )
    }

    override fun length(): Long {
        return reader.length()
    }

    override fun writeAt(pos: Long, buffer: ByteArray) {
        writeAt(pos, buffer, 0 , buffer.size)
    }

    override fun writeAt(pos: Long, buffer: ByteArray, start: Int, offset: Int) {
        super.writeAt(pos, buffer, start, offset)
        _cacheReader=null
    }

}