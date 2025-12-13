package com.qapsoft.io

open class ByteArrayStreamReader(private val bytes:ByteArray):BinaryStreamReader {
    override fun readAt(pos: Long, buffer: ByteArray, start: Int, offset: Int): Int {
        if (pos < 0 || pos > length()) {
            throw ArrayIndexOutOfBoundsException("Position $pos is out of bounds (length: ${length()})")
        }
        
        val available = length() - pos
        if (available == 0L) return 0
        
        val readable = minOf(available, offset.toLong()).toInt()
        val destCapacity = buffer.size - start
        val actualRead = minOf(readable, destCapacity)

        if (actualRead <= 0) return 0

        System.arraycopy(bytes, pos.toInt(), buffer, start, actualRead)
        
        return actualRead
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        return readAt(pos, buffer, 0, buffer.size)
    }

    override fun length(): Long {
        return bytes.size.toLong()
    }
}