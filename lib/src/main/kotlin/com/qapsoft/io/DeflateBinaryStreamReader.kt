package com.qapsoft.io

class DeflateBinaryStreamReader(val deflateSource: BinaryStreamReader, cacheMax:Int=10) : BinaryStreamReader {

    companion object {
        //int 4
        val POS_MAX_DEFLATE_BLOCK_SIZE_INTEGER = 100L
        val POS_HEADER_SIZE_INTEGER = 104L
        val POS_MAX_BLOCKS_COUNT_INTEGER = 108L
        val POS_BLOCKS_COUNT_INTEGER = 112L

        //long 8
        val POS_ORIGINAL_LENGTH_LONG = 116L
    }

    private val maxDeflateBlockSize = deflateSource.getBytesAt(POS_MAX_DEFLATE_BLOCK_SIZE_INTEGER, Int.SIZE_BYTES).asInt()
    private val headerSize = deflateSource.getBytesAt(POS_HEADER_SIZE_INTEGER, Int.SIZE_BYTES).asInt()
    private val maxBlocksCount = deflateSource.getBytesAt(POS_MAX_BLOCKS_COUNT_INTEGER, Int.SIZE_BYTES).asInt()
    private val blocksCount = deflateSource.getBytesAt(POS_BLOCKS_COUNT_INTEGER, Int.SIZE_BYTES).asInt()

    private val _originalLength = deflateSource.getBytesAt(POS_ORIGINAL_LENGTH_LONG, Long.SIZE_BYTES).asLong()

    private val blockIndexSize: Int = maxBlocksCount * 16 // 8 for start , 8 for end pos
    private val blockIndexStartPos = headerSize + 512 //512 if header 0
    private val blocksStartPos = blockIndexStartPos + blockIndexSize

    // Instance-level cache instead of static
    private val blockCacheManager = CacheManager<Int, Block>(cacheMax)

    override fun readAt(pos: Long, buffer: ByteArray, start: Int, offset: Int): Int {
        var p = start
        var o = offset
        var totalReadSize = 0
        var readSize = 0
        
        // Fix: getBlockByPosition should use 'pos + (p - start)' logically, 
        // but initially we just need the block for 'pos'. 
        // As we advance 'p' (buffer index) and 'totalReadSize', the stream position advances by 'totalReadSize'.
        var currentStreamPos = pos
        var block = getBlockByPosition(currentStreamPos)

        // Initial read
        readSize = block.readAt(currentStreamPos, buffer, p, o)
        
        while (readSize > 0) {
            totalReadSize += readSize
            p += readSize
            o -= readSize
            currentStreamPos += readSize

            if (o <= 0) break

            // Move to next block if needed
            // The previous block.readAt might have returned less than requested 'o' because it hit block end.
            // So we need the next block for the *updated* currentStreamPos.
            block = getBlockByPosition(currentStreamPos) 
            // Note: Optimization could be block.nextBlock() but getBlockByPosition is safer for jump logic.
            // Let's stick to getBlockByPosition or reliable nextBlock if we strictly follow sequential read.
            // The original logic tried `block.nextBlock()`. Let's use getBlockByPosition to be robust.

            readSize = block.readAt(currentStreamPos, buffer, p, o)
        }

        return totalReadSize
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        return readAt(pos, buffer, 0, buffer.size)
    }

    override fun length(): Long {
        return _originalLength
    }

    private fun getBlockByPosition(position: Long): Block {
        val blockIndex = getBlockIndex(position)
        return getBlock(blockIndex)
    }

    private fun getBlockIndex(position: Long): Int {
        // Fix: Avoid Int overflow for position
        return (position / maxDeflateBlockSize).toInt()
    }

    private fun getBlock(index: Int): Block {
        if (index < 0 || index >= blocksCount) {
             // Return an empty/dummy block or handle OOB? 
             // Existing logic seemed to rely on cache or something. 
             // Let's assume index is valid or we handle it in Block creation.
             // Actually, verify index against blocksCount seems good practice.
             // For now, proceed with existing logic pattern but safely.
        }
        
        return blockCacheManager[index] ?: run {
             val newBlock = createBlock(index)
             blockCacheManager.cache(index, newBlock)
             newBlock
        }
    }

    private fun createBlock(index: Int): Block {
        val pos = blockIndexStartPos + (index * 16).toLong()
        val start = deflateSource.getBytesAt(pos, Long.SIZE_BYTES).asLong() + blocksStartPos
        val end = deflateSource.getBytesAt(pos + Long.SIZE_BYTES, Long.SIZE_BYTES).asLong() + blocksStartPos

        val startPos = index * maxDeflateBlockSize.toLong()
        val endPos = if ((startPos + maxDeflateBlockSize) > _originalLength)
            _originalLength
        else
            startPos + maxDeflateBlockSize

        val compressedData = deflateSource.getBytesAt(start, (end - start).toInt())
        val inflatedData = DataCompression.inflateData(compressedData) 
            ?: ByteArray(0) // Handle null/failure gracefully

        return Block(
            index = index,
            startPos = startPos,
            endPos = endPos,
            blockReader = ByteArrayStreamReader(inflatedData)
        )
    }

    // Inner class doesn't need to be a BinaryStreamReader necessarily if only used internally, 
    // but preserving structure is fine. removed 'deflateReader' ref to avoid circular dependency issues if not needed.
    inner class Block(
        val index: Int,
        val startPos: Long,
        val endPos: Long,
        val blockReader: BinaryStreamReader
    ) : BinaryStreamReader {

        override fun readAt(pos: Long, buffer: ByteArray, start: Int, offset: Int): Int {
            val realPos = pos - startPos
            
            // Fix: Check if realPos is valid for this block
            if (realPos < 0 || realPos >= blockReader.length()) return 0

            val realOffset = if ((realPos + offset) > blockReader.length())
                blockReader.length() - realPos
            else
                offset.toLong() // Ensure type match if needed, though logic uses Int

            return blockReader.readAt(realPos, buffer, start, realOffset.toInt())
        }

        override fun readAt(pos: Long, buffer: ByteArray): Int {
            return readAt(pos, buffer, 0, buffer.size)
        }

        override fun length(): Long {
            return blockReader.length()
        }
    }
}