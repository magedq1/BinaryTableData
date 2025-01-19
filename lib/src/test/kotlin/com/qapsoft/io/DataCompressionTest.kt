package com.qapsoft.io

import org.junit.jupiter.api.Assertions.*

import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class DataCompressionTest {

    lateinit var testData:ByteArray
    @BeforeEach
    fun setUp() {

        //get random testData
        val a = ByteArrayStreamReaderTest()
        a.setUp()
        testData=a.streamReader.toByteArray()
    }

    @Test
    fun deflateData() {
        assertDoesNotThrow {
            DataCompression.deflateData(testData)
        }
    }

    @Test
    fun inflateData() {
        assertDoesNotThrow {
            val deflated = DataCompression.deflateData(testData)
            val inflated = DataCompression.inflateData(deflated)
            assertNotEquals(deflated.size, inflated?.size)
            assertArrayEquals(
                testData,
                inflated
            )
        }
    }
}