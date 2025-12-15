package com.qapsoft.io

import java.io.Closeable
import java.io.FileDescriptor
import java.io.FileInputStream
import java.io.IOException
import java.nio.ByteBuffer
import java.nio.channels.FileChannel

open class BinaryStreamImpFileDescriptor(val fileDescriptor: FileDescriptor) : BinaryStream, Closeable {

    private val channel: FileChannel = FileInputStream(fileDescriptor).channel

    override fun readAt(pos: Long, buffer: ByteArray, start: Int, offset: Int): Int {
        val byteBuffer = ByteBuffer.wrap(buffer, start, offset)
        return channel.read(byteBuffer, pos)
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        return readAt(pos, buffer, 0, buffer.size)
    }

    override fun length(): Long {
        return channel.size()
    }

    override fun writeAt(pos: Long, buffer: ByteArray, start: Int, offset: Int) {
        throw IOException("This stream is read-only!")
    }

    override fun writeAt(pos: Long, buffer: ByteArray) {
        writeAt(pos, buffer, 0, buffer.size)
    }

    override fun close() {
        channel.close()
    }
}
