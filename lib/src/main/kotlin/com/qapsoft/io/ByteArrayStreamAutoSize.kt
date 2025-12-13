package com.qapsoft.io

class ByteArrayStreamAutoSize(
    blockSize: Int = 1024
) : ByteArrayStreamWriterAutoSize(blockSize), BinaryStream {

    override fun readAt(
        pos: Long,
        buffer: ByteArray,
        start: Int,
        offset: Int
    ): Int {
        val available = minOf(offset.toLong(), length - pos)
        if (available <= 0) return 0

        var remaining = available.toInt()
        var dstPos = start
        var blockIndex = (pos / blockSize).toInt()
        var blockOffset = (pos % blockSize).toInt()

        while (remaining > 0) {
            if (blockIndex >= blocks.size) break
            val block = blocks[blockIndex]
            val readable = minOf(blockSize - blockOffset, remaining)

            System.arraycopy(
                block,
                blockOffset,
                buffer,
                dstPos,
                readable
            )

            remaining -= readable
            dstPos += readable
            blockIndex++
            blockOffset = 0
        }
        return available.toInt() - remaining
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int =
        readAt(pos, buffer, 0, buffer.size)

    override fun length(): Long = length
}
