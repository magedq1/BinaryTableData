package com.qapsoft.io

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertDoesNotThrow
import kotlin.random.Random

class ByteArrayStreamReaderTest {

    lateinit var streamReader:ByteArrayStreamReader
    @BeforeEach
    fun setUp() {
        val random = Random.Default
        val source = ByteArray(100*1024) //100KB
        for(i in 0 until source.size){
            source[i] = random.nextInt(255).toByte()
        }
        streamReader = ByteArrayStreamReader(source)
    }

    @Test
    fun readAt() {
        assertDoesNotThrow {
            val random = Random.Default
            val end = 10*1024 //random 5K
            for(i in 0 until end){
                val pos = random.nextInt(100*1024-1)//100KB-1 guess why ^_^

                val a= streamReader.getBytesAt(pos.toLong(),1)[0]
                val b= streamReader.bytes[pos]
                assertEquals(a,b)
            }
        }
    }


    @Test
    fun length() {
        assertEquals(streamReader.length().toInt(), streamReader.bytes.size)
    }


}