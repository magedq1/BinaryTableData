package com.qapsoft.io

import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import kotlin.random.Random

class DeflateBinaryStreamWriterKtTest {

    companion object{
        fun generateRandomByteArray(size:Int, start:Int=0, end:Int=255):ByteArray{
            val random = Random.Default
            val result = ByteArray(size)
            for(i in 0 until result.size){
                result[i] = random.nextInt(start, end).toByte()
            }
            return result
        }
    }
    @Test
    fun deflate() {
        val reader = ByteArrayStreamReader(
            generateRandomByteArray(
                size= 10*1024*1024, //10MB
                end = 64
            )
        )
        val writer = ByteArrayStreamWriterAutoSize()
        reader.Deflate(
            writer
        )

        val inflate = DeflateBinaryStreamReader(ByteArrayStreamReader(writer.toByteArray()))
        val inflatedBytes = inflate.getBytesAt(0, inflate.length().toInt())
        Assertions.assertArrayEquals(inflatedBytes, reader.toByteArray())

    }


}