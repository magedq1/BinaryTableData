package com.qapsoft.io

open class ByteArrayStreamWriterAutoSize(
    val blockSize: Int = 1024
) : BinaryStreamWriter {

    protected var length: Long = 0
    protected val blocks = ArrayList<ByteArray>()

    override fun writeAt(pos: Long, buffer: ByteArray, start: Int, offset: Int) {
        var remaining = offset
        var bufferPos = start
        var blockIndex = (pos / blockSize).toInt()
        var blockOffset = (pos % blockSize).toInt()

        val endPos = pos + offset
        val maxBlockIndex = ((endPos - 1) / blockSize).toInt()

        ensureCapacity(maxBlockIndex, endPos)

        while (remaining > 0) {
            val block = blocks[blockIndex]
            val writable = minOf(blockSize - blockOffset, remaining)

            System.arraycopy(
                buffer,
                bufferPos,
                block,
                blockOffset,
                writable
            )

            bufferPos += writable
            remaining -= writable
            blockIndex++
            blockOffset = 0
        }
    }

    override fun writeAt(pos: Long, buffer: ByteArray) {
        writeAt(pos, buffer, 0, buffer.size)
    }

    private fun ensureCapacity(blockIndex: Int, newLength: Long) {
        synchronized(this) {
            while (blocks.size <= blockIndex) {
                blocks.add(ByteArray(blockSize))
            }
            if (length < newLength) {
                length = newLength
            }
        }
    }

    fun toByteArray(): ByteArray {
        val result = ByteArray(length.toInt())
        var destPos = 0

        for (i in blocks.indices) {
            val block = blocks[i]
            val copySize = minOf(blockSize, result.size - destPos)
            if (copySize <= 0) break

            System.arraycopy(block, 0, result, destPos, copySize)
            destPos += copySize
        }
        return result
    }
}
