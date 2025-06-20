package com.qapsoft.io

import org.junit.jupiter.api.Assertions.*

import org.junit.jupiter.api.Test

class BinaryTableHeaderKtTest {

    @Test
    fun asBinaryColumnList() {
        val cols = listOf(
            BinaryColumn.Int("_id"),
            BinaryColumn.String("name"),
            BinaryColumn.String("mobile"),
            BinaryColumn.Double("amount"),
            BinaryColumn.Long("date"),
            BinaryColumn.Raw("finger_print",10),
        )
        val encoded = cols.encodedAsByteArray()

        println(cols.first().columnType)
        val decoded = encoded.asBinaryColumnList()
        println(decoded.first().columnType)
        assertArrayEquals(
            cols.toTypedArray(),
            decoded.toTypedArray()
        )
    }
}