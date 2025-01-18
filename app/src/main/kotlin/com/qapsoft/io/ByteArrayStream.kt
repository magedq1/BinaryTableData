package com.qapsoft.io

class ByteArrayStream(private val bytes:ByteArray):ByteArrayStreamReader(bytes), BinaryStreamWriter {
    var _length:Long = 0
    override fun writeAt(pos: Long, buffer: ByteArray, start: Int, offset: Int) {
        if(pos>bytes.size)
            throw ArrayIndexOutOfBoundsException()
        val o:Int = if(pos+offset>bytes.size)
                    bytes.size-pos.toInt()
                else
                    offset
        val end = start+o
        var p = pos.toInt()
        for(i in start until end){
            bytes[p++] = buffer[i]
        }
        if(p>_length)
            _length=p.toLong()
    }

    override fun writeAt(pos: Long, buffer: ByteArray) {
        writeAt(pos, buffer, 0 ,buffer.size)
    }

    override fun length(): Long {
        return _length
    }
}