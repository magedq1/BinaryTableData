package com.qapsoft.io

import org.junit.jupiter.api.Assertions.*

import org.junit.jupiter.api.Test
import kotlin.random.Random

class ByteArrayStreamWriterAutoSizeTest {

    @Test
    fun writeAt() {
        assertDoesNotThrow {
            val rand = Random.Default
            val writer = ByteArrayStreamWriterAutoSize(rand.nextInt(1024*10)) //10K
            val randomArray = DeflateBinaryStreamWriterKtTest.generateRandomByteArray(
                rand.nextInt(0, 1024*500)
            )
            writer.writeAt(0, randomArray)
        }
    }

    @Test
    fun toByteArray() {
        val rand = Random.Default
        val writer = ByteArrayStreamWriterAutoSize(rand.nextInt(1024*10)) //10K
        val randomArray = DeflateBinaryStreamWriterKtTest.generateRandomByteArray(
            rand.nextInt(0, 1024*500)
        )
        writer.writeAt(0, randomArray)
        assertArrayEquals(
            randomArray,
            writer.toByteArray()
        )
    }
}