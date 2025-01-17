package com.qapsoft.io

class DeflateBinaryStreamReader(val reader: BinaryStreamReader):BinaryStreamReader {

    //int 4
    private val POS_MAX_DEFLATE_BLOCK_SIZE_INTEGER      = 100L
    private val POS_HEADER_SIZE_INTEGER                 = 104L
    private val POS_MAX_BLOCKS_COUNT_INTEGER            = 108L
    private val POS_BLOCKS_COUNT_INTEGER                = 112L

    //long 8
    private val POS_ORIGINAL_LENGTH_LONG                = 116L

    private val maxDeflateBlockSize = reader.getBytesAt(POS_MAX_DEFLATE_BLOCK_SIZE_INTEGER, Int.SIZE_BYTES).asInt()
    private val headerSize = reader.getBytesAt(POS_HEADER_SIZE_INTEGER, Int.SIZE_BYTES).asInt()
    private val maxBlocksCount = reader.getBytesAt(POS_MAX_BLOCKS_COUNT_INTEGER, Int.SIZE_BYTES).asInt()
    private val blocksCount = reader.getBytesAt(POS_BLOCKS_COUNT_INTEGER, Int.SIZE_BYTES).asInt()

    private val _originalLength = reader.getBytesAt(POS_ORIGINAL_LENGTH_LONG, Long.SIZE_BYTES).asLong()

    private val blockIndexSize:Int = maxBlocksCount*8 // 4 for start , 4 for end pos

    override fun readAt(pos: Long, buffer: ByteArray, start: Int, offset: Int): Int {
        val blocks = getBlocks(pos, offset)

        return offset
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        return readAt(pos, buffer, 0, buffer.size)
    }

    override fun length(): Long {
        return _originalLength
    }
    fun getBlocks(pos: Long, offset: Int): List<BinaryStreamReader>{
        return emptyList()
    }

    data class Block(
        private val startPos: Long,
        private val endPos:Long,
        val reader:DeflateBinaryStreamReader
    )

    class InflatedBlock(
        val reader:BinaryStreamReader //mostly be ByteArrayStreamReader
    ):BinaryStreamReader{
        override fun readAt(pos: Long, buffer: ByteArray, start: Int, offset: Int): Int {
            return reader.readAt(pos, buffer, start, offset)
        }

        override fun readAt(pos: Long, buffer: ByteArray): Int {
            return readAt(pos, buffer, 0 ,buffer.size)
        }

        override fun length(): Long {
            return end-start
        }

    }
}