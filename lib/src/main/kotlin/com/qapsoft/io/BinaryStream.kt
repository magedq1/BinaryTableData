package com.qapsoft.io

import java.io.IOException

interface BinaryStream:BinaryStreamReader, BinaryStreamWriter {

}
fun BinaryStreamReader.toStream(throwWriteError: Boolean=true):BinaryStream{
    return object :BinaryStream{
        override fun readAt(
            pos: Long,
            buffer: ByteArray,
            start: Int,
            offset: Int
        ): Int {
            return this@toStream.readAt(pos, buffer, start, offset)
        }

        override fun readAt(pos: Long, buffer: ByteArray): Int {
            return this@toStream.readAt(pos, buffer)
        }

        override fun length(): Long {
            return this@toStream.length()
        }

        override fun writeAt(
            pos: Long,
            buffer: ByteArray,
            start: Int,
            offset: Int
        ) {
            throw IOException("read only stream!")
        }

        override fun writeAt(pos: Long, buffer: ByteArray) {
            throw IOException("read only stream!")
        }

    }
}