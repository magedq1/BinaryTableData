package com.qapsoft.io

class DeflateBinaryStreamReader(reader: BinaryStreamReader):BinaryStreamReader {

    companion object{
        //int 4
        val POS_MAX_DEFLATE_BLOCK_SIZE_INTEGER      = 100L
        val POS_HEADER_SIZE_INTEGER                 = 104L
        val POS_MAX_BLOCKS_COUNT_INTEGER            = 108L
        val POS_BLOCKS_COUNT_INTEGER                = 112L

        //long 8
        val POS_ORIGINAL_LENGTH_LONG                = 116L
    }

    private val maxDeflateBlockSize = reader.getBytesAt(POS_MAX_DEFLATE_BLOCK_SIZE_INTEGER, Int.SIZE_BYTES).asInt()
    private val headerSize = reader.getBytesAt(POS_HEADER_SIZE_INTEGER, Int.SIZE_BYTES).asInt()
    private val maxBlocksCount = reader.getBytesAt(POS_MAX_BLOCKS_COUNT_INTEGER, Int.SIZE_BYTES).asInt()
    private val blocksCount = reader.getBytesAt(POS_BLOCKS_COUNT_INTEGER, Int.SIZE_BYTES).asInt()

    private val _originalLength = reader.getBytesAt(POS_ORIGINAL_LENGTH_LONG, Long.SIZE_BYTES).asLong()

    private val blockIndexSize:Int = maxBlocksCount*16 // 8 for start , 8 for end pos
    private val blockIndexStartPos = headerSize+512 //512 if header 0
    private val blocksStartPos = blockIndexStartPos+blockIndexSize

    override fun readAt(pos: Long, buffer: ByteArray, start: Int, offset: Int): Int {
        var p = start
        var o = offset
        var totalReadSize = 0
        var readSize = 0
        var block = Block.getBlockByPosition(pos, this)
        readSize=block.readAt(pos+p, buffer, p, o)
        while (readSize>0){

            totalReadSize+=readSize
            p+=readSize
            o-=readSize
            block=block.nextBlock()?:return totalReadSize
            readSize=block.readAt(pos+p, buffer, p, o)
        }

        return totalReadSize
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        return readAt(pos, buffer, 0, buffer.size)
    }

    override fun length(): Long {
        return _originalLength
    }


    data class Block(
        private val index: Int,
        private val startPos: Long,
        private val endPos:Long,
        val blockReader:BinaryStreamReader,
        val deflateReader: DeflateBinaryStreamReader
    ):BinaryStreamReader{
        companion object{
//            private val blockCacheManager = CacheManager<Block>(50)
            var blockCacheManager = CacheManager<DeflateBinaryStreamReader,CacheManager<Int,Block>>(10)
            private var reader:DeflateBinaryStreamReader?=null

            fun blockCounts (reader:DeflateBinaryStreamReader?=null) = (reader?:this.reader)?.blocksCount?:0

            fun setDefaultReader(reader:DeflateBinaryStreamReader){
                this.reader = reader
            }
            operator fun get(index: Int, reader:DeflateBinaryStreamReader?=null):Block{
                val validReader:DeflateBinaryStreamReader = reader?:this.reader!!
                if(!blockCacheManager.containsKey(validReader)){
                    blockCacheManager.cache(validReader, CacheManager(50))
                }
                val cache = blockCacheManager[validReader]!!
                return cache[index]?: cache.cache(
                    key = index,
                    value = getBlock(index, validReader)
                )
            }
            fun getBlockByPosition(position: Long, reader:DeflateBinaryStreamReader?=null):Block{
                val blockIndex = getBlockIndex(position, reader)
                return get(blockIndex, reader)
            }

            fun getBlockIndex(position: Long, reader: DeflateBinaryStreamReader?=null): Int {
                val index:Int = position.toInt()/(reader?:this.reader!!).maxDeflateBlockSize
                return index
            }

            private fun getBlock(index:Int,reader:DeflateBinaryStreamReader):Block{
                return reader.run {
                    val pos = blockIndexStartPos+(index*16).toLong()
                    val start = getBytesAt(pos,Long.SIZE_BYTES).asLong()+blocksStartPos
                    val end = getBytesAt(pos+Long.SIZE_BYTES,Long.SIZE_BYTES).asLong()+blocksStartPos

                    val startPos = index*maxDeflateBlockSize.toLong()
                    val endPos = if((startPos+maxDeflateBlockSize)>_originalLength)
                                    _originalLength
                                else
                                    startPos+maxDeflateBlockSize
                    Block(
                        index = index,
                        startPos = startPos,
                        endPos = endPos,
                        blockReader = ByteArrayStreamReader(
                            DataCompression.inflateData(
                                getBytesAt(start, (end-start).toInt())
                            )!!
                        ),
                        deflateReader = reader
                    )
                }
            }
        }

        fun nextBlock():Block?{
            val nextIndex = index+1
            if(nextIndex>=deflateReader.blocksCount)
                return null
            return get(nextIndex, deflateReader)
        }
        override fun readAt(pos: Long, buffer: ByteArray, start: Int, offset: Int): Int {
            val realPos = pos-startPos
            val realOffset = if((realPos+offset)>blockReader.length())
                                blockReader.length()-realPos
                            else
                                offset
            return blockReader.readAt(realPos, buffer, start, realOffset.toInt())
        }

        override fun readAt(pos: Long, buffer: ByteArray): Int {
            return readAt(pos,  buffer,0, buffer.size)
        }

        override fun length(): Long {
            return blockReader.length()
        }

    }


}