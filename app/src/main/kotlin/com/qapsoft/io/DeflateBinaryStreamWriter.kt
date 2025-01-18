package com.qapsoft.io

fun BinaryStreamReader.Deflate(output:BinaryStreamWriter, blockSize:Int=8192, headerSize:Int=0){
    synchronized(this){
        val maxBlocksCount:Int = (length()/blockSize).toInt()+1
        val usedBlocksCount= if(length().mod(length()/blockSize)==0L)
                                maxBlocksCount-1
                            else
                                maxBlocksCount
        output.writeAt(DeflateBinaryStreamReader.POS_MAX_BLOCKS_COUNT_INTEGER, maxBlocksCount.toByteArray())
        output.writeAt(DeflateBinaryStreamReader.POS_BLOCKS_COUNT_INTEGER, usedBlocksCount.toByteArray())
        output.writeAt(DeflateBinaryStreamReader.POS_MAX_DEFLATE_BLOCK_SIZE_INTEGER, blockSize.toByteArray())
        output.writeAt(DeflateBinaryStreamReader.POS_HEADER_SIZE_INTEGER, headerSize.toByteArray())

        output.writeAt(DeflateBinaryStreamReader.POS_ORIGINAL_LENGTH_LONG, length().toByteArray())

        val blockIndexSize:Int = maxBlocksCount*16 // 8 for start , 8 for end pos
        val blockIndexStartPos = headerSize+512 //512 if header 0
        val blocksStartPos = blockIndexStartPos+blockIndexSize

        val byteBuffer = ByteArray(blockSize)
        for(i in 0 until usedBlocksCount){
            val realInputPos = i*blockSize
            val readSize = readAt(realInputPos.toLong(), buffer = byteBuffer)
            val compressedBytes = DataCompression.deflateData(byteBuffer, 0 , readSize)

            //write compressed block
            val realOutputPos = blocksStartPos+(i*blockSize)
            output.writeAt(realOutputPos.toLong(), compressedBytes)

            //write block index
            val realBlockIndexPos = blockIndexStartPos+(i*16)
            output.writeAt(realBlockIndexPos.toLong(), (i*blockSize).toLong().toByteArray())
            output.writeAt(realBlockIndexPos.toLong()+8, (i*blockSize+compressedBytes.size).toLong().toByteArray())
        }
    }
}