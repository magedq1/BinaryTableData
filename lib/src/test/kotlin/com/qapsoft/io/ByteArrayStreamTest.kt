package com.qapsoft.io

import org.junit.jupiter.api.Assertions.assertArrayEquals
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.BeforeEach

import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertDoesNotThrow
import org.junit.jupiter.api.assertThrows
import kotlin.random.Random

class ByteArrayStreamTest {

    lateinit var testBytes:ByteArray
    @BeforeEach
    fun setUp(){
        testBytes = ByteArrayStreamReaderTest()
            .also {
                it.setUp()
            }
            .streamReader.toByteArray()

    }

    @Test
    fun writeAt() {
        assertDoesNotThrow {
            val stream = ByteArrayStream(ByteArray(testBytes.size))
            stream.writeAt(0, testBytes)
            assertArrayEquals(testBytes, stream.toByteArray())
        }
    }
    @Test
    fun writeRandom(){
        val stream = ByteArrayStream(ByteArray(testBytes.size))
        val random = Random.Default
        repeat(500){
            val pos = random.nextInt(testBytes.size)
            val offset = random.nextInt(pos, testBytes.size)-pos
            stream.writeAt(pos.toLong(), testBytes.copyOfRange(pos, pos+offset))
            assertArrayEquals(testBytes.copyOfRange(pos, pos+offset), stream.getBytesAt(pos.toLong(),offset))
        }
    }
    @Test
    fun writeOver(){
        var stream = ByteArrayStream(ByteArray(testBytes.size))
        assertThrows<ArrayIndexOutOfBoundsException> {
            stream.writeAt(testBytes.size.toLong()+1, ByteArray(50))
        }
        val b = ByteArray(500)
            .also { b->
                repeat(500){i->
                    b[i] = 255.toByte()
                }
            }
        stream = ByteArrayStream(ByteArray(testBytes.size))
        stream.writeAt(testBytes.size.toLong(), b)
        assertArrayEquals(stream.toByteArray(), ByteArray(testBytes.size))


        stream = ByteArrayStream(ByteArray(testBytes.size))
        stream.writeAt(testBytes.size.toLong()-250, b)
        assertEquals(250, stream.toByteArray().filter {
            it!=0.toByte()
        }.size)

    }
}