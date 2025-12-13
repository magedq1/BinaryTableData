package com.qapsoft.io

import java.io.InputStream

class SubBinaryStreamReader(val source: BinaryStreamReader, val startPos:Long, val endPos:Long=source.length()):
    BinaryStreamReader {

    private val subLength = endPos - startPos + 1

    init {
        require(startPos >= 0) { "start must be >= 0" }
        require(endPos >= startPos) { "end must be >= start" }
        require(endPos < source.length()) { "end exceeds source length" }
    }

    override fun readAt(
        pos: Long,
        buffer: ByteArray,
        start: Int,
        offset: Int
    ): Int {
        if (pos < 0 || pos >= subLength) return -1

        val maxReadable = minOf(
            offset.toLong(),
            subLength - pos
        ).toInt()

        return source.readAt(
            this.startPos + pos,
            buffer,
            start,
            maxReadable
        )
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        if (pos < 0 || pos >= subLength) return -1

        val maxReadable = minOf(
            buffer.size.toLong(),
            subLength - pos
        ).toInt()

        return source.readAt(
            this.startPos + pos,
            buffer,
            0,
            maxReadable
        )
    }

    override fun length(): Long = subLength
}

fun BinaryStreamReader.subReader(startPos:Long, endPos:Long=length()):SubBinaryStreamReader{
    return SubBinaryStreamReader(this, startPos, endPos)
}
fun BinaryStreamReader.toByteArray(): ByteArray{
    return getBytesAt(0, length().toInt())
}
fun BinaryStreamReader.asInputStream(): InputStream {
    val reader = this
    return object : InputStream() {

        private var position: Long = 0
        private val length: Long = reader.length()

        override fun read(): Int {
            if (position >= length) return -1

            val buffer = ByteArray(1)
            val bytesRead = reader.readAt(position, buffer)

            return if (bytesRead <= 0) {
                -1
            } else {
                position++
                buffer[0].toInt() and 0xFF
            }
        }

        override fun read(b: ByteArray, off: Int, len: Int): Int {
            if (position >= length) return -1
            if (off < 0 || len < 0 || off + len > b.size) {
                throw IndexOutOfBoundsException()
            }

            val bytesToRead = minOf(len.toLong(), length - position).toInt()
            val bytesRead = reader.readAt(position, b, off, bytesToRead)

            if (bytesRead > 0) {
                position += bytesRead
            }

            return if (bytesRead <= 0) -1 else bytesRead
        }

        override fun skip(n: Long): Long {
            val skipped = minOf(n, length - position)
            position += skipped
            return skipped
        }

        override fun available(): Int {
            return minOf((length - position), Int.MAX_VALUE.toLong()).toInt()
        }
    }
}