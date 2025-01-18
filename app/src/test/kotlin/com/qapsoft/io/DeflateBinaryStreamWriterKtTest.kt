package com.qapsoft.io

import org.junit.jupiter.api.Test
import kotlin.random.Random

class DeflateBinaryStreamWriterKtTest {

    @Test
    fun deflate() {
        val reader = ByteArrayStreamReader(
            generateRandomByteArray(
                10*1024*1024 //10MB
            )
        )
        val writer = ByteArrayStream(ByteArray(reader.length().toInt()))
        reader.Deflate(writer)

    }
    fun generateRandomByteArray(size:Int):ByteArray{
        val random = Random.Default
        val result = ByteArray(size)
        for(i in 0 until result.size){
            result[i] = random.nextInt(255).toByte()
        }
        return result
    }

}