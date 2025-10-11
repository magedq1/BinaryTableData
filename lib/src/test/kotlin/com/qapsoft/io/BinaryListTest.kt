package com.qapsoft.io

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows

class BinaryListTest {



    @Test
    fun all() {
        val list = BinaryList(
            TestStream()
        )
        assert(list.size==0)
        assert(
            list.add {
                it.write("maged".toByteArray())
            }==0
        )
        assert(list.size==1)
        assert(
            list.add {
                it.write("ahmed".toByteArray())
            }==1
        )
        assert(list.size==2)

        assertArrayEquals("maged".toByteArray(), list.get(0)?.readBytes())
        assertArrayEquals("ahmed".toByteArray(), list.get(1)?.readBytes())
        assertThrows<IndexOutOfBoundsException> {
            list.get(2)
        }

    }

}
private class TestStream: ByteArrayStreamWriterAutoSize(),  BinaryStream{
    override fun readAt(
        pos: Long,
        buffer: ByteArray,
        start: Int,
        offset: Int
    ): Int {
        return ByteArrayStreamReader(toByteArray()).readAt(
            pos, buffer, start, offset
        )
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        return ByteArrayStreamReader(toByteArray()).readAt(
            pos, buffer
        )
    }

    override fun length(): Long {
        return ByteArrayStreamReader(toByteArray()).length()
    }

}