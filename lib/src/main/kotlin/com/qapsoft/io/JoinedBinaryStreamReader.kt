package com.qapsoft.io

import java.util.Random

class JoinedBinaryStreamReader(
    private val splitReaders : List<BinaryStreamReader>
):BinaryStreamReader {
    protected val lengths = mutableListOf<Long>()
    private val fullLength:Long
    init {
        var lastLength = 0L
        for(s in splitReaders){
            lengths.add(s.length()+lastLength)
            lastLength=s.length()+lastLength
        }
        fullLength = lastLength
    }
    override fun readAt(
        pos: Long,
        buffer: ByteArray,
        start: Int,
        offset: Int
    ): Int {
        if (pos >= fullLength) {
            return 0 // Nothing to read at or after the end
        }

        var readerIndex = -1
        for (i in lengths.indices) {
            if (pos < lengths[i]) {
                readerIndex = i
                break
            }
        }

        if (readerIndex == -1) {
            return 0 // Should not happen if pos < fullLength, but as a safeguard
        }

        val previousLength = if (readerIndex == 0) 0L else lengths[readerIndex - 1]
        var currentPosInSplit = pos - previousLength

        var totalBytesRead = 0
        var currentBufferOffset = start
        var remainingBytesToRead = offset

        for (i in readerIndex until splitReaders.size) {
            if (remainingBytesToRead == 0) {
                break
            }

            val currentReader = splitReaders[i]
            
            val readableBytesInCurrentSplit = currentReader.length() - currentPosInSplit
            val bytesToReadFromThisSplit = minOf(remainingBytesToRead.toLong(), readableBytesInCurrentSplit).toInt()

            if (bytesToReadFromThisSplit <= 0) {
                currentPosInSplit = 0
                continue
            }
            
            val bytesRead = currentReader.readAt(currentPosInSplit, buffer, currentBufferOffset, bytesToReadFromThisSplit)

            if (bytesRead > 0) {
                totalBytesRead += bytesRead
                currentBufferOffset += bytesRead
                remainingBytesToRead -= bytesRead
            }

            if (bytesRead < bytesToReadFromThisSplit) {
                // Reached the end of this reader, or an error occurred. Stop reading.
                break
            }
            
            // For the next reader, start reading from the beginning.
            currentPosInSplit = 0
        }

        return totalBytesRead
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        return readAt(pos, buffer, 0, buffer.size)
    }

    override fun length(): Long {
        return fullLength
    }
}