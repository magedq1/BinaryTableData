package com.qapsoft.io

open class ByteArrayStreamWriterAutoSize(private val blockSize:Int=1024):BinaryStreamWriter {
    private var _length:Long =0
    private val blocks = mutableListOf<ByteArray>()


    override fun writeAt(pos: Long, buffer: ByteArray, start: Int, offset: Int) {
        var blockNum:Int = pos.toInt()/blockSize
        synchronized(this){
            var remainOffset = offset
            var startBlockWritePos = (pos-(blockNum*blockSize)).toInt()
            var bufferCursor = start
            while (remainOffset>0){
                val bytes = getBlock(blockNum++)
                val bytesToWrite:Int =
                    if(remainOffset>(blockSize-startBlockWritePos))
                        (blockSize-startBlockWritePos)
                    else
                        remainOffset

                for(i in startBlockWritePos until blockSize){
                    if(bufferCursor>=offset)
                        continue
                    bytes[i] = buffer[bufferCursor++]
                }

                remainOffset-=bytesToWrite
                startBlockWritePos=0
            }
            if(_length<(pos+offset))
                _length = pos+offset
        }
    }

    private fun getBlock(blockNum: Int):ByteArray {
        if(blockNum>=blocks.size){
            for(i in blocks.size..blockNum){
                blocks.add(ByteArray(blockSize))
            }
        }
        return blocks[blockNum]
    }

    override fun writeAt(pos: Long, buffer: ByteArray) {
        writeAt(pos, buffer, 0 , buffer.size)
    }

    fun toByteArray():ByteArray{
        val res = ByteArray(_length.toInt())
        var endBlock = blockSize
        for(i in 0 until blocks.size){
            if(i==(blocks.size-1)){
                endBlock= _length.toInt()-(i*blockSize)
            }
            val block = getBlock(i)
            block.copyInto(
                destination = res,
                destinationOffset = i*blockSize,
                endIndex = endBlock
            )
        }
        return res
    }
}