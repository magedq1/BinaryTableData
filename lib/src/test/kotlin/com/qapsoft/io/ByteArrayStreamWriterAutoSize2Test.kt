package com.qapsoft.io

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test
import kotlin.random.Random

class ByteArrayStreamWriterAutoSize2Test {

    @Test
    fun testWriteExceedingInitialCapacity() {
        val blockSize = 100
        val writer = ByteArrayStreamWriterAutoSize(blockSize)
        
        // Data spanning more than one block (e.g., 2.5 blocks)
        val dataSize = 250
        val data = ByteArray(dataSize) { it.toByte() }
        
        assertDoesNotThrow {
            writer.writeAt(0, data)
        }
        
        // Verify content
        val result = writer.toByteArray()
        assertArrayEquals(data, result)
    }

    @Test
    fun testWriteAtOffsetExceedingCapacity() {
        val blockSize = 100
        val writer = ByteArrayStreamWriterAutoSize(blockSize)
        
        val data = ByteArray(50) { it.toByte() }
        
        // Write at a position that requires expanding beyond initial blocks
        // writing at 150 (in 2nd block)
        assertDoesNotThrow {
            writer.writeAt(150, data)
        }
    }
}
