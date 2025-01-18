package com.qapsoft.io

import org.junit.jupiter.api.Assertions.assertArrayEquals
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertDoesNotThrow
import org.junit.jupiter.api.assertThrows
import kotlin.random.Random

class ByteArrayStreamReaderTest {

    lateinit var streamReader:ByteArrayStreamReader
    lateinit var source:ByteArray
    @BeforeEach
    fun setUp() {
        val random = Random.Default
        source = ByteArray(100*1024) //100KB
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
                val b= source[pos]
                assertEquals(a,b)
            }
        }
    }

    @Test
    fun readRandom(){
        val random = Random.Default
        repeat(500){
            val pos = random.nextInt(source.size)
            val offset = random.nextInt(pos, source.size)-pos
            assertArrayEquals(source.copyOfRange(pos, pos+offset), streamReader.getBytesAt(pos.toLong(),offset))
        }
    }
    @Test
    fun readOverOffset(){
        assertThrows<ArrayIndexOutOfBoundsException> {
            streamReader.getBytesAt(source.size.toLong()+1,50)
        }
        assert( streamReader.readAt(source.size.toLong(), ByteArray(50))==0)
        assert( streamReader.readAt(source.size.toLong()-30, ByteArray(50))==30)
    }

}