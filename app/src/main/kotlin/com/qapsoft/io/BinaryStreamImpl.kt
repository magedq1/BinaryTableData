package com.qapsoft.io

import java.io.Closeable
import java.io.File
import java.io.RandomAccessFile
import java.nio.MappedByteBuffer
import java.nio.channels.FileChannel

class BinaryStreamImpl(file: File, size:Int):BinaryStream, Closeable{
    val channel: FileChannel
    val mappedByteBuffer: MappedByteBuffer
    init {
        channel = RandomAccessFile(file, "rw").channel
        mappedByteBuffer = channel.map(FileChannel.MapMode.READ_WRITE, 0, size.toLong())
        channel.read(mappedByteBuffer)
    }
    override fun readAt(pos: Long,buffer:ByteArray, start: Int, offset: Int): Int {
        this.mappedByteBuffer.position(pos.toInt())
        this.mappedByteBuffer.get(buffer, start, offset)
        return offset
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        return readAt(pos,buffer,0 , buffer.size)
    }

    override fun writeAt(pos: Long, buffer:ByteArray,start: Int, offset: Int) {
        this.mappedByteBuffer.position(pos.toInt())
        this.mappedByteBuffer.put(buffer, start, offset)
    }

    override fun writeAt(pos: Long, buffer: ByteArray) {
        writeAt(pos,buffer, 0 ,buffer.size)
    }

    override fun close() {
        mappedByteBuffer.force()
        channel.close()
    }
}